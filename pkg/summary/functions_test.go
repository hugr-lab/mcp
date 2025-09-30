package summary

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hugr-lab/mcp/pkg/pool"
)

func TestFunction(t *testing.T) {
	// load schema
	schema, err := loadSchema()
	if err != nil {
		t.Fatal(err)
	}

	m := schema.Module("tf.iot")
	if m == nil {
		t.Fatal("module not found")
	}

	//find function
	function := schema.Function("tf.iot.road_part_sectors_car_count_current")
	if function == nil {
		t.Fatal("function not found")
	}

	input, err := prepareFunctionInput(t.Context(), schema, function)
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Function input (%d): %+v", len(b), string(b))

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
	summary, err := s.SummarizeFunction(t.Context(), schema, function)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Function summary: %+v", summary)
}
