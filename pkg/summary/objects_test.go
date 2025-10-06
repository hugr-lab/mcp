package summary

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hugr-lab/mcp/pkg/pool"
)

func TestObjects(t *testing.T) {
	// load schema
	schema, err := loadSchema()
	if err != nil {
		t.Fatal(err)
	}

	// select table
	table := schema.Table("tf.digital_twin.tf_road_parts")
	if table == nil {
		t.Fatal("table not found ")
	}

	input, err := prepareDataObjectInput(schema, table)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Table Input: %+v", input)
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Table Input JSON length: %d", len(b)/1024)

	s := New(
		pool.Config{
			Timeout:        600 * time.Second,
			MaxConnections: 10,
			Provider:       pool.ProviderCustom,
			Model:          "openai/gpt-oss-20b",
			//Model:   "mistral/devstral-small-2507",
			//Model:   "qwen/qwen3-4b-2507",
			BaseUrl: "http://localhost:1234/v1",
			ApiKey:  "test_api_key",
		},
	)

	out, err := s.SummarizeDataObject(t.Context(), schema, table)

	if err != nil {
		t.Fatal(err)
	}

	b, err = json.Marshal(out)

	t.Logf("Table Summary: (%d) %s", len(b)/1024, string(b))
}
