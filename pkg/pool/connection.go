package pool

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	llms "github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderCustom    ProviderType = "custom" // openAI compatible model
)

type Config struct {
	Timeout        time.Duration
	MaxConnections int
	Provider       ProviderType
	Model          string
	ApiKey         string
	BaseUrl        string
	// Tools
}

type Pool struct {
	size int
	used atomic.Int32
	cfg  Config
}

func New(config Config) *Pool {
	return &Pool{
		size: config.MaxConnections,
		used: atomic.Int32{},
		cfg:  config,
	}
}

var ErrUnknownProvider = errors.New("unknown provider")

func (pool *Pool) Connection(ctx context.Context, opts ...llms.CallOption) (c *Connection, err error) {
	// wait available connection (used and size of pool)
	for pool.used.Load() >= int32(pool.size) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	var m llms.Model
	switch pool.cfg.Provider {
	case ProviderOpenAI:
		// Initialize OpenAI connection
		m, err = openai.New(
			openai.WithToken(pool.cfg.ApiKey),
			openai.WithModel(pool.cfg.Model),
		)
	case ProviderAnthropic:
		// Initialize Anthropic connection
		m, err = anthropic.New(
			anthropic.WithToken(pool.cfg.ApiKey),
			anthropic.WithModel(pool.cfg.Model),
		)
	case ProviderCustom:
		// Initialize Custom connection
		m, err = openai.New(
			openai.WithBaseURL(pool.cfg.BaseUrl),
			openai.WithToken(pool.cfg.ApiKey),
			openai.WithModel(pool.cfg.Model),
		)
	default:
		return nil, ErrUnknownProvider
	}

	if err != nil {
		return nil, err
	}

	return &Connection{
		llm:     m,
		timeout: pool.cfg.Timeout,
		closeFunc: func() error {
			pool.used.Add(-1)
			return nil
		},
	}, nil
}

type Connection struct {
	llm     llms.Model
	timeout time.Duration
	opts    []llms.CallOption

	closeFunc func() error
}

func (c *Connection) Close() error {
	if c.closeFunc != nil {
		err := c.closeFunc()
		if err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
		// Reset the close function to prevent double closing
		c.closeFunc = nil
	}
	return nil
}

type SummarizationTask struct {
	SystemPrompt       string
	UserPromptTemplate string
	Data               any
	MaxTokens          int
	Temperature        float64
}

// Use the llm model to generate a summary based on the task
func (c *Connection) Summarize(ctx context.Context, task *SummarizationTask) (string, error) {

	tmpl := template.New("userPrompt")
	tmpl.Parse(task.UserPromptTemplate)
	var builder = &strings.Builder{}
	err := tmpl.Execute(builder, task.Data)
	if err != nil {
		return "", fmt.Errorf("failed to execute user prompt template: %w", err)
	}

	msgs := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, task.SystemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, builder.String()),
	}
	opts := append([]llms.CallOption{}, c.opts...)
	if task.MaxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(task.MaxTokens))
	}
	if task.Temperature > 0 {
		opts = append(opts, llms.WithTemperature(task.Temperature))
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	out, err := c.llm.GenerateContent(ctx, msgs, opts...)
	if err != nil {
		return "", err
	}
	return out.Choices[0].Content, nil
}

func (c *Connection) Call(ctx context.Context, prompt string, maxTokens int, temperature float64) (string, error) {
	msgs := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	opts := []llms.CallOption{
		llms.WithMaxTokens(maxTokens),
		llms.WithTemperature(temperature),
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	out, err := c.llm.GenerateContent(ctx, msgs, opts...)
	if err != nil {
		return "", err
	}
	return out.Choices[0].Content, nil
}
