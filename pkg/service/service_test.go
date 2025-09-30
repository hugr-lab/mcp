package service

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/hugr-lab/mcp/pkg/pool"
	hugr "github.com/hugr-lab/query-engine"
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
		URL:          os.Getenv("HUGR_IPC_URL"),
		Secret:       os.Getenv("HUGR_SECRET"),
		SecretHeader: os.Getenv("HUGR_SECRET_HEADER"),
		TTL:          testGetEnvDuration("HUGR_CACHE_TTL", 60*time.Second),
		Indexer: indexer.Config{
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
