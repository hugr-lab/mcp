package pool

import (
	"context"
	"testing"
	"time"
)

func TestSummarize(t *testing.T) {
	pool := New(Config{
		Timeout:        60 * time.Second,
		MaxConnections: 10,
		Provider:       ProviderCustom,
		Model:          "openai/gpt-oss-20b",
		BaseUrl:        "http://localhost:1234/v1",
		ApiKey:         "test_api_key",
	})

	conn, err := pool.Connection(context.Background())
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	defer conn.Close()

	task := &SummarizationTask{
		SystemPrompt:       "You are a helpful assistant. Return the text exactly as it is.",
		UserPromptTemplate: "Summarize the following text: {{.Text}}",
		Data:               map[string]interface{}{"Text": "This is a test."},
		MaxTokens:          100,
		Temperature:        0.7,
	}

	result, err := conn.Summarize(context.Background(), task)
	if err != nil {
		t.Fatalf("failed to summarize: %v", err)
	}

	expected := "This is a test."
	if result != expected {
		t.Errorf("unexpected result: got %q, want %q", result, expected)
	}
}
