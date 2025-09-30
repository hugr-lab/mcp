package indexer

import "testing"

func TestService_SearchModuleFunctions(t *testing.T) {
	s := New(testConfig, testHugr)
	res, err := s.SearchModuleFunctions(t.Context(), &SearchFunctionsRequest{
		Module:            "",
		Query:             "load data source",
		TopK:              30,
		MinScore:          0.2,
		IncludeSubModules: true,
		IncludeMutations:  false,
	})
	if err != nil {
		t.Fatalf("failed to search module functions: %v", err)
	}
	t.Logf("found %d functions:", res.Total)
	for _, fn := range res.Items {
		t.Logf("- %s: %s (score: %.4f)", fn.Name, fn.Description, fn.Score)
	}
}
