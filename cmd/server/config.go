package main

import (
	"github.com/hugr-lab/mcp/pkg/indexer"
	"github.com/hugr-lab/mcp/pkg/pool"
	"github.com/hugr-lab/mcp/pkg/service"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func init() {
	initEnvs()
}

func initEnvs() {
	// Initialize environment variables
	_ = godotenv.Load()
	viper.SetDefault("HUGR_IPC_URL", "")
	viper.SetDefault("BIND", ":14000")
	viper.AutomaticEnv()
}

type Config struct {
	service.Config
	Bind string
}

func config() Config {
	return Config{
		Config: service.Config{
			URL:          viper.GetString("HUGR_IPC_URL"),
			Secret:       viper.GetString("HUGR_SECRET"),
			SecretHeader: viper.GetString("HUGR_SECRET_HEADER"),
			TTL:          viper.GetDuration("HUGR_CACHE_TTL"),
			Indexer: indexer.Config{
				Path:       viper.GetString("INDEXER_DATA_SOURCE_PATH"),
				VectorSize: viper.GetInt("INDEXER_VECTOR_SIZE"),
				ReadOnly:   viper.GetBool("INDEXER_READ_ONLY"),
				// Embeddings
				EmbeddingsEnabled: viper.GetBool("EMBEDDINGS_ENABLED"),
				EmbeddingModel:    viper.GetString("EMBEDDINGS_MODEL"),
				// Summarization
				SummarizeSchema: viper.GetBool("SUMMARIZE_SCHEMA"),
				Summarize: pool.Config{
					Timeout:        viper.GetDuration("SUMMARIZE_TIMEOUT"),
					MaxConnections: viper.GetInt("SUMMARIZE_MAX_CONNECTIONS"),
					Provider:       pool.ProviderType(viper.GetString("SUMMARIZE_PROVIDER")),
					Model:          viper.GetString("SUMMARIZE_MODEL"),
					BaseUrl:        viper.GetString("SUMMARIZE_BASE_URL"),
					ApiKey:         viper.GetString("SUMMARIZE_API_KEY"),
				},
				CacheTTL: viper.GetDuration("INDEXER_CACHE_TTL"),
			},
		},
		Bind: viper.GetString("BIND"),
	}
}
