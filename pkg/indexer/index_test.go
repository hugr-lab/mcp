package indexer

import (
	"testing"
)

func TestServiceIndexFields(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.IndexFields(t.Context(), false); err != nil {
		t.Fatalf("failed to index fields: %v", err)
	}
}

func TestServiceIndexField(t *testing.T) {
	s := New(testConfig, testHugr)
	res, err := s.h.Query(t.Context(), `query {
		core {
			mcp {
				fields(
					filter: { 
						type_name: { eq: "_module_owm_function" }
					}
					order_by: [
						{field: "type_name"}
						{field: "name"}
					]
				) {
					name
					type_name
					description
				}
			}
		}
	}`, nil)
	if err != nil {
		t.Fatalf("failed to query fields for indexing: %v", err)
	}
	defer res.Close()
	if res.Err() != nil {
		t.Fatalf("failed to query fields for indexing: %v", res.Err())
	}
	var fields []Field
	if err := res.ScanData("core.mcp.fields", &fields); err != nil {
		t.Fatalf("failed to decode fields: %v", err)
	}
	for _, f := range fields {
		if err := s.IndexField(t.Context(), f.Name, f.TypeName, f.Description); err != nil {
			t.Fatalf("failed to index field: %v", err)
		}
	}
}

func TestServiceIndexTypes(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.IndexTypes(t.Context(), false); err != nil {
		t.Fatalf("failed to index types: %v", err)
	}
}

func TestServiceIndexModules(t *testing.T) {
	s := New(testConfig, testHugr)

	err := s.IndexModules(t.Context(), true)
	if err != nil {
		t.Fatalf("failed to index modules: %v", err)
	}
}

func TestServiceIndexDataSources(t *testing.T) {
	s := New(testConfig, testHugr)
	err := s.IndexDataSources(t.Context(), true)
	if err != nil {
		t.Fatalf("failed to index data sources: %v", err)
	}
}
