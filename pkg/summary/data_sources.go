package summary

import (
	"context"
	"encoding/json"
	"errors"

	_ "embed"

	"github.com/hugr-lab/mcp/pkg/pool"
)

//go:embed templates/data_source.tmpl
var dataSourceTemplate string

type DataSource struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ReadOnly    bool             `json:"read_only"`
	AsModule    bool             `json:"as_module"`
	Tables      []DataSourceItem `json:"tables"`
	Views       []DataSourceItem `json:"views"`
	Functions   []DataSourceItem `json:"functions"`
	Submodules  []DataSourceItem `json:"submodules"`
}

type DataSourceItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DataSourceSummary struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}

func (s *Service) SummarizeDataSource(ctx context.Context, ds DataSource) (*DataSourceSummary, error) {
	data, err := prepareDataSourceInput(ds)
	if err != nil {
		return nil, err
	}

	c, err := s.pool.Connection(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	resp, err := c.Summarize(ctx, &pool.SummarizationTask{
		SystemPrompt:       systemPrompt,
		UserPromptTemplate: dataSourceTemplate,
		Data:               data,
		MaxTokens:          4096,
		Temperature:        0.3,
	})
	if err != nil {
		return nil, err
	}

	var summary DataSourceSummary
	err = json.Unmarshal([]byte(resp), &summary)
	if err != nil {
		return nil, errors.Join(ErrSummarizationOutputFormat, err)
	}

	return &summary, nil
}

func prepareDataSourceInput(ds DataSource) (*DataSourceDescribeInput, error) {
	input := DataSourceDescribeInput{
		Name:        ds.Name,
		Description: ds.Description,
		Type:        "unknown",
		AsModule:    ds.AsModule,
		ReadOnly:    ds.ReadOnly,
	}

	b, err := json.Marshal(ds.Tables)
	if err != nil {
		return nil, err
	}
	input.TablesJSON = string(b)

	b, err = json.Marshal(ds.Views)
	if err != nil {
		return nil, err
	}
	input.ViewsJSON = string(b)

	b, err = json.Marshal(ds.Functions)
	if err != nil {
		return nil, err
	}
	input.FunctionsJSON = string(b)

	b, err = json.Marshal(ds.Submodules)
	if err != nil {
		return nil, err
	}
	input.SubmodulesJSON = string(b)

	return &input, nil
}

type DataSourceDescribeInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	AsModule    bool   `json:"as_module"`
	ReadOnly    bool   `json:"read_only"`

	TablesJSON     string `json:"tables"`
	ViewsJSON      string `json:"views"`
	FunctionsJSON  string `json:"functions"`
	SubmodulesJSON string `json:"modules"`
}
