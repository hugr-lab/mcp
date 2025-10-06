package summary

import "errors"

var ErrSummarizationOutputFormat = errors.New("invalid output format")

// Optional lightweight context objects
type DataSourceContext struct {
	Name        string `json:"name,omitempty"`
	SummaryText string `json:"summary_text,omitempty"`
}

type ModuleContext struct {
	Name     string `json:"name"`
	Overview string `json:"overview,omitempty"`
}
