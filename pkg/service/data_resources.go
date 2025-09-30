package service

// The tool to execute GraphQL queries and save results as a resources for use in the analysis chain

// Resource types:
// - the data parts for each query in the GraphQL query can be JSON, table (parquet).
// - JQ transformations results - JSON
// - Summarization results - text

// Resource metadata:
type DataResource struct {
	ID          string             `json:"id"`
	SessionID   string             `json:"session_id,omitempty"`
	Description string             `json:"description,omitempty"`
	Query       string             `json:"query,omitempty"`
	Parts       []DataResourcePart `json:"parts,omitempty"`
	CreatedAt   int64              `json:"created_at,omitempty"`
	TTL         int64              `json:"ttl,omitempty"` // seconds to live
}

type DataResourcePart struct {
	Path     string `json:"path,omitempty"`
	Type     string `json:"type,omitempty"`   // query_result, jq_transform, summarization
	Format   string `json:"format,omitempty"` // json, parquet
	Rows     int64  `json:"rows,omitempty"`   // for tabular data
	Size     int64  `json:"size,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

// Resources will be stored in the storage with metadata in the DB
// Resources can be accessible via presigned URLs for a limited time
