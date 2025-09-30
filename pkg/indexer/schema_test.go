package indexer

import (
	"testing"

	hugr "github.com/hugr-lab/query-engine"
)

func Test_fetchSchema(t *testing.T) {
	h := hugr.NewClient("http://localhost:15003/ipc")
	s := New(Config{}, h)
	schema, err := s.fetchSchema(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}
	t.Logf("fetched %d schema types", len(schema.Types))
}

func Test_fetchSummary(t *testing.T) {
	h := hugr.NewClient("http://localhost:15003/ipc")
	s := New(Config{}, h)
	summary, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}
	t.Logf("fetched summary: %+v", summary)
}
