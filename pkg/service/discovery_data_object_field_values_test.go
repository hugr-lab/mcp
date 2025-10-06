package service

import "testing"

func TestDiscoveryDataObjectFieldValues(t *testing.T) {
	s := New(testConfig)

	// get data object queries info
	tqq, err := s.indexer.DataObjectQueriesInfo(t.Context(), "tf_dictionary_types")
	if err != nil {
		t.Fatalf("failed to get type info: %v", err)
	}
	if tqq == nil {
		t.Fatal("data object not found")
	}

	stats, err := s.queryDataObjectFieldStats(t.Context(), tqq, schemaDataObjectFieldValuesInput{
		ObjectName:     "tf_dictionary_types",
		FieldName:      "name",
		Limit:          20,
		CalculateStats: true,
		Args: map[string]any{
			"language": "en",
		},
	})
	if err != nil {
		t.Fatalf("failed to query field stats: %v", err)
	}

	t.Logf("Stats: %+v", stats)

}
