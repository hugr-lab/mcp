package indexer

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	tests := []struct {
		name, path string
	}{
		{
			name: "duckdb",
			path: "test_duckdb.db",
		},
		{
			name: "postgres",
			path: os.Getenv("TEST_PG_DB"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitDB(t.Context(), tt.path, 768)
			if err != nil {
				t.Fatalf("failed to init db: %v", err)
			}
		})
	}
}
