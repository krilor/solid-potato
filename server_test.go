package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krilor/solid-potato/db"
)

type statsQuerier struct {
	db.Querier
	rows []db.ListTransactionStatsRow
}

func (q statsQuerier) ListTransactionStats(context.Context) ([]db.ListTransactionStatsRow, error) {
	return q.rows, nil
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
