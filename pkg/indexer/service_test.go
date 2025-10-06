package indexer

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hugr-lab/mcp/pkg/pool"
	hugr "github.com/hugr-lab/query-engine"
	"github.com/hugr-lab/query-engine/pkg/db"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	testHugr   *hugr.Client
	testConfig Config
)

func TestMain(m *testing.M) {
	hurl := os.Getenv("HUGR_IPC_URL")
	if hurl == "" {
		log.Fatal("HUGR_IPC_URL is not set")
	}
	testHugr = hugr.NewClient(hurl, hugr.WithTimeout(90*time.Second))

	testConfig = Config{
		Path:              os.Getenv("INDEXER_DATA_SOURCE_PATH"),
		ReadOnly:          testGetEnvBool("INDEXER_READ_ONLY", false),
		VectorSize:        testGetEnvInt("INDEXER_VECTOR_SIZE", 768),
		CacheTTL:          testGetEnvDuration("INDEXER_CACHE_TTL", 60*time.Second),
		EmbeddingsEnabled: testGetEnvBool("EMBEDDINGS_ENABLED", false),
		EmbeddingModel:    os.Getenv("EMBEDDINGS_MODEL"),
		SummarizeSchema:   testGetEnvBool("SUMMARIZE_SCHEMA", true),
		Summarize: pool.Config{
			Timeout:        testGetEnvDuration("SUMMARIZE_TIMEOUT", 900*time.Second),
			MaxConnections: testGetEnvInt("SUMMARIZE_MAX_CONNECTIONS", 2),
			Provider:       pool.ProviderType(os.Getenv("SUMMARIZE_PROVIDER")),
			Model:          os.Getenv("SUMMARIZE_MODEL"),
			BaseUrl:        os.Getenv("SUMMARIZE_BASE_URL"),
			ApiKey:         os.Getenv("SUMMARIZE_API_KEY"),
		},
	}

	os.Exit(m.Run())
}

func testGetEnvInt(name string, defaultValue int) int {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}
	var iv int
	iv, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return iv
}

func testGetEnvBool(name string, defaultValue bool) bool {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}
	return v == "true" || v == "1"
}

func testGetEnvDuration(name string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
}

func TestServiceInit(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}
}

func TestSchemaParse(t *testing.T) {
	schema, err := db.ParseSQLScriptTemplate(db.SDBAttachedPostgres, hschema, dbInitParams{
		DBVersion:         dbVersion,
		VectorSize:        testConfig.VectorSize,
		EmbeddingsEnabled: testConfig.EmbeddingsEnabled,
		EmbeddingModel:    testConfig.EmbeddingModel,
	})
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	t.Logf("parsed schema: %s", schema)
}

func TestServiceLoad(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.fillBaseSchema(t.Context()); err != nil {
		t.Fatalf("failed to load service: %v", err)
	}
}

func TestAddField(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.AddField(t.Context(), Field{
		TypeName:    "Unknown",
		Name:        "test_field",
		Description: "A test field",
		Type:        "String",
	}); err != nil {
		t.Fatalf("failed to add field: %v", err)
	}
}

func TestService_ModulesRank(t *testing.T) {
	s := New(testConfig, testHugr)

	mm, err := s.SearchModules(t.Context(), "aggregated roads information", 5, 0)
	if err != nil {
		t.Fatalf("failed to rank modules : %v", err)
	}
	t.Logf("ranked modules: %+v", mm)
}

func TestService_DataSourcesRank(t *testing.T) {
	s := New(testConfig, testHugr)

	mm, err := s.SearchDataSources(t.Context(), "aggregated roads information", 5, 0)
	if err != nil {
		t.Fatalf("failed to rank data sources : %v", err)
	}
	t.Logf("ranked data sources: %+v", mm)
}

func TestService_addDataObjects(t *testing.T) {
	s := New(testConfig, testHugr)

	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	doo := meta.DataObjects()
	if len(doo) == 0 {
		t.Fatal("no data objects found")
	}
	t.Logf("found %d data objects", len(doo))

	res, err := s.h.Query(t.Context(), `mutation {
		core {
			mcp {
				delete_data_object_queries { success }
				delete_data_objects { success }
			}
		}
	}`, nil)
	if err != nil {
		t.Fatalf("query clear types: %v", err)
	}
	defer res.Close()
	if res.Err() != nil {
		t.Errorf("query clear types: %v", res.Err())
	}

	for _, do := range doo {
		m := meta.Module(do.Module)
		if m == nil || m.QueryType == "" {
			t.Errorf("module %q not found for data object %q", do.Module, do.Name)
			continue
		}
		object := DataObject{
			Name:           do.Name,
			FilterTypeName: do.FilterType,
		}
		if do.Arguments != nil {
			object.ArgsTypeName = do.Arguments.Type
		}
		for _, q := range do.Queries {
			object.Queries = append(object.Queries, DataObjectQuery{
				Name:      q.Name,
				QueryType: string(q.Type),
				QueryRoot: m.QueryType,
			})
		}

		err := s.addDataObject(t.Context(), object)
		if err != nil {
			t.Errorf("failed to add data object %q: %v", do.Name, err)
		}
	}

}

func TestService_updateArgsIsNotNull(t *testing.T) {
	s := New(testConfig, testHugr)
	ctx := t.Context()
	// 1. fetch schema types
	schema, err := s.fetchSchema(ctx)
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	// 3. Add all types from schema
	var aa []Argument
	am := map[string]struct{}{}
	for _, st := range schema.Types {
		fields := st.Fields
		if st.Kind == string(ast.InputObject) {
			fields = st.InputFields
		}
		for _, f := range fields {
			for _, a := range f.Args {
				key := fmt.Sprintf("%s.%s.%s", st.Name, f.Name, a.Name)
				if _, ok := am[key]; ok {
					t.Fatalf("duplicate argument %q in hugr schema", key)
				}
				am[key] = struct{}{}
				if !a.Type.IsNotNull() {
					continue
				}
				arg := Argument{
					Name:        a.Name,
					FieldName:   f.Name,
					TypeName:    st.Name,
					Description: a.Description,
					Type:        a.Type.TypeName(),
					IsList:      a.Type.IsList(),
					IsNotNull:   a.Type.IsNotNull(),
				}
				aa = append(aa, arg)
			}
		}
	}

	// 5. arguments
	for _, a := range aa {
		err := testUpdateIsNotNullArg(ctx, s.h, a)
		if err != nil {
			t.Fatalf("failed to add argument %q: %v", a.Name, err)
		}
	}
}

func testUpdateIsNotNullArg(ctx context.Context, h *hugr.Client, a Argument) error {
	res, err := h.Query(ctx, `mutation updateArgIsNotNull($typeName: String!, $fieldName: String!, $argName: String!, $isNotNull: Boolean!) {
		core {
			mcp {
				update_arguments(
					filter: { 
						type_name: { eq: $typeName }
						field_name: { eq: $fieldName }
						name: { eq: $argName }
					}
					data: {
						is_non_null: $isNotNull
					}
				) {
					success
				}
			}
		}
	}`, map[string]any{
		"typeName":  a.TypeName,
		"fieldName": a.FieldName,
		"argName":   a.Name,
		"isNotNull": a.IsNotNull,
	})
	if err != nil {
		return fmt.Errorf("failed to update argument is_not_null: %w", err)
	}
	defer res.Close()
	if res.Err() != nil {
		return fmt.Errorf("failed to update argument is_not_null: %w", res.Err())
	}
	return nil
}

func TestService_loadScalarDescription(t *testing.T) {
	s := New(testConfig, testHugr)
	ctx := t.Context()
	// 1. fetch schema types
	schema, err := s.fetchSchema(ctx)
	if err != nil {
		t.Fatalf("failed to fetch schema: %v", err)
	}

	// update all scalars
	for _, st := range schema.Types {
		if st.Kind != string(ast.Scalar) {
			continue
		}
		err := s.UpdateTypeDescription(ctx, st.Name, st.Description, "", false)
		if err != nil {
			t.Errorf("failed to update scalar description for %q: %v", st.Name, err)
		}
	}
}
