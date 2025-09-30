package summary

import (
	"github.com/hugr-lab/mcp/pkg/pool"
)

// LLM Wrapper to perform summarization tasks for Hugr GraphQL schema descriptions

type Service struct {
	pool *pool.Pool
}

func New(llm pool.Config) *Service {
	return &Service{
		pool: pool.New(llm),
	}
}

// Initialize any necessary resources or dependencies for the service
func (s *Service) Init() error {

	return nil
}
