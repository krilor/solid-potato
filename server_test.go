package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/krilor/solid-potato/db"
)

func TestParseListTransactionsParams(t *testing.T) {
	ts := func(s string) *time.Time {
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return &parsed
	}
	status := func(s TransactionStatus) *TransactionStatus { return &s }

	tests := []struct {
		name    string
		query   string
		want    ListTransactionsParams
		wantErr string
	}{
		{name: "no parameters", query: ""},
		{
			name:  "from and to",
			query: "from=2024-01-01T00:00:00Z&to=2024-02-01T12:30:00%2B01:00",
			want:  ListTransactionsParams{From: ts("2024-01-01T00:00:00Z"), To: ts("2024-02-01T12:30:00+01:00")},
		},
		{name: "only from", query: "from=2024-01-01T00:00:00Z", want: ListTransactionsParams{From: ts("2024-01-01T00:00:00Z")}},
		{name: "only to", query: "to=2024-01-01T00:00:00Z", want: ListTransactionsParams{To: ts("2024-01-01T00:00:00Z")}},
		{name: "status approved", query: "status=APPROVED", want: ListTransactionsParams{Status: status(APPROVED)}},
		{name: "status pending", query: "status=PENDING", want: ListTransactionsParams{Status: status(PENDING)}},
		{name: "status rejected", query: "status=REJECTED", want: ListTransactionsParams{Status: status(REJECTED)}},
		{
			name:  "all parameters",
			query: "from=2024-01-01T00:00:00Z&to=2024-02-01T00:00:00Z&status=REJECTED",
			want: ListTransactionsParams{
				From:   ts("2024-01-01T00:00:00Z"),
				To:     ts("2024-02-01T00:00:00Z"),
				Status: status(REJECTED),
			},
		},
		{name: "invalid from", query: "from=yesterday", wantErr: `"from" must be an RFC 3339 date-time`},
		{name: "date only from", query: "from=2024-01-01", wantErr: `"from" must be an RFC 3339 date-time`},
		{name: "empty from", query: "from=", wantErr: `"from" must be an RFC 3339 date-time`},
		{name: "invalid to", query: "to=nope", wantErr: `"to" must be an RFC 3339 date-time`},
		{name: "unknown status", query: "status=UNKNOWN", wantErr: `"status" must be a single value`},
		{name: "lowercase status", query: "status=approved", wantErr: `"status" must be a single value`},
		{name: "empty status", query: "status=", wantErr: `"status" must be a single value`},
		{name: "repeated status", query: "status=APPROVED&status=PENDING", wantErr: `"status" must be a single value`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/transactions?"+tt.query, nil)

			got, err := parseListTransactionsParams(request)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !timePtrEqual(got.From, tt.want.From) {
				t.Errorf("From = %v, want %v", got.From, tt.want.From)
			}
			if !timePtrEqual(got.To, tt.want.To) {
				t.Errorf("To = %v, want %v", got.To, tt.want.To)
			}
			if (got.Status == nil) != (tt.want.Status == nil) || (got.Status != nil && *got.Status != *tt.want.Status) {
				t.Errorf("Status = %v, want %v", got.Status, tt.want.Status)
			}
		})
	}
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

func TestListOrganizationStats(t *testing.T) {
	queries := statsQuerier{rows: []db.ListTransactionStatsRow{{
		OrgNumber:   "313227502",
		OrgName:     "Example AS",
		SumQuantity: 1375.22,
		NumApproved: 10,
		NumRejected: 2,
		NumPending:  1,
	}}}
	handler := newServer(queries)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/stats/org", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got, want := recorder.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got, want := recorder.Body.String(), "[{\"numApproved\":10,\"numPending\":1,\"numRejected\":2,\"orgName\":\"Example AS\",\"orgNumber\":\"313227502\",\"sumQuantity\":1375.22}]\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

type statsQuerier struct {
	db.Querier
	rows []db.ListTransactionStatsRow
}

func (q statsQuerier) ListTransactionStats(context.Context) ([]db.ListTransactionStatsRow, error) {
	return q.rows, nil
}

func TestSanitizedQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "empty", query: "", want: ""},
		{name: "sorted", query: "status=APPROVED&from=2025-01-01T00:00:00Z", want: "from=2025-01-01T00%3A00%3A00Z&status=APPROVED"},
		{name: "control characters are escaped", query: "status=A%0AB%0D", want: "status=A%0AB%0D"},
		{name: "long values are truncated", query: "status=" + strings.Repeat("a", 500), want: "status=" + strings.Repeat("a", maxLoggedQueryLen) + "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/transactions?"+tt.query, nil)
			got := sanitizedQuery(r)
			if strings.ContainsAny(got, "\r\n") {
				t.Errorf("query contains raw control characters: %q", got)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
