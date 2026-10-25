package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/krilor/solid-potato/db"
)

//go:embed openapi.json
var openAPISpec []byte

type server struct {
	queries db.Querier
}

// newServer returns the API http.Handler, wrapped with request logging and panic recovery.
func newServer(queries db.Querier) http.Handler {
	s := &server{queries: queries}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /transactions", s.listTransactions)
	mux.HandleFunc("GET /transactions/{reference}", s.readTransaction)
	mux.HandleFunc("GET /stats/org", s.listOrganizationStats)
	mux.HandleFunc("GET /openapi.json", serveOpenAPI)

	return recoverer(logRequests(mux))
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(openAPISpec); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func (s *server) listTransactions(w http.ResponseWriter, r *http.Request) {
	params, err := parseListTransactionsParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request parameters: "+err.Error()+".")
		return
	}

	var arg db.ListTransactionsParams
	if params.From != nil {
		arg.From = pgtype.Timestamptz{Time: *params.From, Valid: true}
	}
	if params.To != nil {
		arg.To = pgtype.Timestamptz{Time: *params.To, Valid: true}
	}
	if params.Status != nil {
		arg.Status = pgtype.Text{String: string(*params.Status), Valid: true}
	}

	rows, err := s.queries.ListTransactions(r.Context(), arg)
	if err != nil {
		internalError(w, r, "list transactions", err)
		return
	}

	resp := make([]TransactionResponse, len(rows))
	for i, row := range rows {
		resp[i] = toTransactionResponse(row.Reference, row.CardNumber, row.CardType, row.TransactionTime, row.OrgNumber, row.OrgName, row.Quantity, row.Status)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) readTransaction(w http.ResponseWriter, r *http.Request) {
	var ref pgtype.UUID
	if err := ref.Scan(r.PathValue("reference")); err != nil || !ref.Valid {
		writeError(w, http.StatusBadRequest, "Invalid transaction reference.")
		return
	}

	row, err := s.queries.ReadTransactionByReference(r.Context(), ref)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "Transaction not found.")
		return
	}
	if err != nil {
		internalError(w, r, "read transaction", err)
		return
	}

	writeJSON(w, http.StatusOK, toTransactionResponse(row.Reference, row.CardNumber, row.CardType, row.TransactionTime, row.OrgNumber, row.OrgName, row.Quantity, row.Status))
}

func (s *server) listOrganizationStats(w http.ResponseWriter, r *http.Request) {
	rows, err := s.queries.ListTransactionStats(r.Context())
	if err != nil {
		internalError(w, r, "list organization stats", err)
		return
	}

	resp := make([]OrganizationStatsResponse, len(rows))
	for i, row := range rows {
		resp[i] = OrganizationStatsResponse{
			OrgNumber:   row.OrgNumber,
			OrgName:     row.OrgName,
			SumQuantity: row.SumQuantity,
			NumApproved: row.NumApproved,
			NumRejected: row.NumRejected,
			NumPending:  row.NumPending,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseListTransactionsParams(r *http.Request) (ListTransactionsParams, error) {
	var params ListTransactionsParams
	q := r.URL.Query()

	for _, p := range []struct {
		name string
		dst  **time.Time
	}{{"from", &params.From}, {"to", &params.To}} {
		if !q.Has(p.name) {
			continue
		}
		t, err := time.Parse(time.RFC3339, q.Get(p.name))
		if err != nil {
			return params, fmt.Errorf("%q must be an RFC 3339 date-time", p.name)
		}
		*p.dst = &t
	}

	if q.Has("status") {
		st := TransactionStatus(q.Get("status"))
		if len(q["status"]) > 1 || !st.Valid() {
			return params, errors.New(`"status" must be a single value of APPROVED, PENDING or REJECTED`)
		}
		params.Status = &st
	}

	return params, nil
}

func toTransactionResponse(reference pgtype.UUID, cardNumber, cardType string, transactionTime pgtype.Timestamptz, orgNumber, orgName string, quantity float32, status string) TransactionResponse {
	return TransactionResponse{
		Reference:       openapi_types.UUID(reference.Bytes),
		CardNumber:      cardNumber,
		CardType:        CardType(cardType),
		TransactionTime: transactionTime.Time,
		OrgNumber:       orgNumber,
		OrgName:         orgName,
		Quantity:        quantity,
		Status:          TransactionStatus(status),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, Error{Message: message})
}

func internalError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	slog.ErrorContext(r.Context(), msg, "error", err, "method", r.Method, "path", r.URL.Path)
	writeError(w, http.StatusInternalServerError, "Internal server error.")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.ErrorContext(r.Context(), "panic in handler", "panic", v, "method", r.Method, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, "Internal server error.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
