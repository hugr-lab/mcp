package indexer

import "testing"

func TestServiceTypeIntrospection(t *testing.T) {
	s := New(testConfig, testHugr)

	out, err := s.TypeIntrospection(t.Context(), "Geometry")
	if err != nil {
		t.Fatalf("failed to get type introspection: %v", err)
	}
	t.Logf("Type Introspection: %+v", out)
}

func TestServiceTypeFieldsIntrospection(t *testing.T) {
	s := New(testConfig, testHugr)

	req := &TypeFieldsRequest{
		TypeName:           "_module_op2023_query",
		IncludeArguments:   true,
		IncludeDescription: true,
		Limit:              10,
		Offset:             0,
	}

	out, err := s.TypeFieldsIntrospection(t.Context(), req)
	if err != nil {
		t.Fatalf("failed to get type fields introspection: %v", err)
	}
	t.Logf("Type Fields Introspection: %+v", out)
}

func TestServiceEnumValuesIntrospection(t *testing.T) {
	s := New(testConfig, testHugr)

	out, err := s.EnumValuesIntrospection(t.Context(), "VectorDistanceType")
	if err != nil {
		t.Fatalf("failed to get enum values introspection: %v", err)
	}
	t.Logf("Enum Values Introspection: %+v", out)
}
