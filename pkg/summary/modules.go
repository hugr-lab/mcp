package summary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	_ "embed"

	"github.com/hugr-lab/mcp/pkg/pool"
	metainfo "github.com/hugr-lab/query-engine/pkg/data-sources/sources/runtime/meta-info"
)

//go:embed templates/module.tmpl
var moduleTemplate string

func (s *Service) SummarizeModule(ctx context.Context, module ModuleInfo) (*ModuleSummary, error) {
	data := ModuleDescribeTemplateData{
		Name:        module.Name,
		Description: module.Description,
	}

	b, err := json.Marshal(module.Tables)
	if err != nil {
		return nil, err
	}
	data.TablesJSON = string(b)
	b, err = json.Marshal(module.Views)
	if err != nil {
		return nil, err
	}
	data.ViewsJSON = string(b)
	b, err = json.Marshal(module.Functions)
	if err != nil {
		return nil, err
	}
	data.FunctionsJSON = string(b)
	b, err = json.Marshal(module.MutFunctions)
	if err != nil {
		return nil, err
	}
	data.MutationFunctionsJSON = string(b)
	b, err = json.Marshal(module.SubModules)
	if err != nil {
		return nil, err
	}
	data.SubmodulesJSON = string(b)

	var dss []DataSourceContext
	for _, ds := range module.DataSources {
		dss = append(dss, DataSourceContext{
			Name:        ds.Name,
			SummaryText: fmt.Sprintf("%s (%s)", ds.Description, ds.Type),
		})
	}
	b, err = json.Marshal(dss)
	if err != nil {
		return nil, err
	}
	data.DataSourceContextsJSON = string(b)

	c, err := s.pool.Connection(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	resp, err := c.Summarize(ctx, &pool.SummarizationTask{
		SystemPrompt:       systemPrompt,
		UserPromptTemplate: moduleTemplate,
		Data:               data,
		MaxTokens:          2096,
		Temperature:        0.3,
	})
	if err != nil {
		return nil, err
	}

	var summary ModuleSummary
	if err := json.Unmarshal([]byte(resp), &summary); err != nil {
		return nil, errors.Join(ErrSummarizationOutputFormat, err)
	}

	return &summary, nil
}

type ModuleDescribeTemplateData struct {
	Name                   string
	Description            string
	TablesJSON             string
	ViewsJSON              string
	FunctionsJSON          string
	MutationFunctionsJSON  string
	SubmodulesJSON         string
	DataSourceContextsJSON string
}

type ModuleInfo struct {
	Name         string
	Description  string
	Tables       map[string]string
	Views        map[string]string
	Functions    map[string]string
	MutFunctions map[string]string
	SubModules   map[string]string
	DataSources  []metainfo.DataSourceInfo
}

type ModuleSummary struct {
	Short           string `json:"short"`
	Long            string `json:"long,omitempty"`
	QueryType       string `json:"query_type,omitempty"`
	MutationType    string `json:"mutation_type,omitempty"`
	FunctionType    string `json:"function_type,omitempty"`
	MutFunctionType string `json:"mutation_function_type,omitempty"`
}
