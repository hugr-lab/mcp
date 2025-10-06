package indexer

import "testing"

func TestService_SearchModuleDataObjects(t *testing.T) {
	s := New(testConfig, testHugr)
	res, err := s.SearchModuleDataObjects(t.Context(), &SearchDataObjectsRequest{
		Module:            "tf",
		Query:             "road parts",
		FieldsQuery:       "attributes road type",
		TopK:              10,
		TopKField:         30,
		MinScore:          0.2,
		MinFieldScore:     0.2,
		IncludeSubModules: true,
	})
	if err != nil {
		t.Fatalf("failed to search module data objects: %v", err)
	}
	t.Logf("found %d data objects:", res.Total)
	for _, do := range res.Items {
		t.Logf("- %s: %s (score: %.4f)", do.Name, do.Description, do.Score)
	}
}
