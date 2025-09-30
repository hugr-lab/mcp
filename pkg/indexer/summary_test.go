package indexer

import (
	"testing"

	"github.com/hugr-lab/mcp/pkg/summary"
	"github.com/hugr-lab/query-engine/pkg/compiler/base"
)

func TestService_summarizeDataObject(t *testing.T) {
	s := New(testConfig, testHugr)

	sum := summary.New(s.c.Summarize)
	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	err = s.SummarizeDataObject(t.Context(), sum, meta, Type{
		Name:     "tf_road_parts",
		Module:   "tf.digital_twin",
		HugrType: base.HugrTypeTable,
	})
	if err != nil {
		t.Fatalf("failed to summarize data object: %v", err)
	}
}

func TestService_summarizeFunction(t *testing.T) {
	s := New(testConfig, testHugr)

	sum := summary.New(s.c.Summarize)
	meta, err := s.fetchSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to fetch summary: %v", err)
	}

	fi := meta.Function("owm.current_weather")
	if fi == nil {
		t.Fatalf("failed to get function info: %v", err)
	}
	m := meta.Module("owm")
	if m == nil {
		t.Fatalf("failed to get module info: %v", err)
	}

	err = s.SummarizeFunction(t.Context(), sum, meta, Field{
		Name:     "current_weather",
		TypeName: m.FunctionType,
	})
	if err != nil {
		t.Fatalf("failed to summarize function: %v", err)
	}
}

func TestService_modulesForSummary(t *testing.T) {
	s := New(testConfig, testHugr)

	modules, err := s.modulesForSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to get modules for summary: %v", err)
	}

	for _, m := range modules {
		t.Logf("Module: %+v", m)
	}
}

func TestService_summarizeModule(t *testing.T) {
	s := New(testConfig, testHugr)

	sum := summary.New(s.c.Summarize)

	err := s.SummarizeModule(t.Context(), sum, "tf2.indicators")
	if err != nil {
		t.Fatalf("failed to summarize module %s: %v", "tf2.indicators", err)
	}
}

func TestService_summarizeModules(t *testing.T) {
	s := New(testConfig, testHugr)

	sum := summary.New(s.c.Summarize)

	mm, err := s.modulesForSummary(t.Context())
	if err != nil {
		t.Fatalf("failed to get modules for summary: %v", err)
	}

	for _, m := range mm {
		err := s.SummarizeModule(t.Context(), sum, m)
		if err != nil {
			t.Logf("failed to summarize module %s: %v", m, err)
		}
	}
}

func TestService_Summarize(t *testing.T) {
	s := New(testConfig, testHugr)
	if err := s.Init(t.Context()); err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	t.Logf("Starting summarization")
	err := s.Summarize(t.Context())
	if err != nil {
		t.Fatalf("failed to summarize: %v", err)
	}
}

func TestService_summarizeDataSource(t *testing.T) {
	s := New(testConfig, testHugr)

	sum := summary.New(s.c.Summarize)

	err := s.SummarizeDataSource(t.Context(), sum, "tf")
	if err != nil {
		t.Fatalf("failed to summarize data source %s: %v", "tf", err)
	}
}
