package summary

import (
	"encoding/json"
	"os"
	"testing"

	hugr "github.com/hugr-lab/query-engine"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
	"github.com/hugr-lab/query-engine/pkg/db"
)

func TestSummary(t *testing.T) {

	h := hugr.NewClient("http://localhost:15003/ipc")
	res, err := h.Query(t.Context(), `query {
		function{
			core{
				meta{
					schema_summary
				}
			}
		}
	}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Close()
	if res.Err() != nil {
		t.Fatal(res.Err())
	}

	part := res.DataPart("function.core.meta.schema_summary")
	if part == nil {
		t.Fatal("missing schema summary")
	}

	data := []byte(*part.(*db.JsonValue))

	// Use the data
	t.Logf("Schema Summary: %+v", len(data)/1024/1024)

	// write bytes to the file
	err = os.WriteFile("schema_summary.json", data, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func loadSchema() (*metainfo.SchemaInfo, error) {
	f, err := os.Open("formatted.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var schema metainfo.SchemaInfo
	if err := json.NewDecoder(f).Decode(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}
