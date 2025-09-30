-- Schema for the indexer service
{{if isPostgres }}
CREATE EXTENSION IF NOT EXISTS vector;
{{end}}

CREATE TABLE version_info (
    version TEXT NOT NULL PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO version_info (version) VALUES ('{{ .DBVersion }}');

CREATE TABLE IF NOT EXISTS types (
    name TEXT NOT NULL PRIMARY KEY,
    description TEXT NOT NULL,
    long_description TEXT NOT NULL,
    kind TEXT NOT NULL,
    hugr_type TEXT NOT NULL,
    module TEXT NOT NULL,
    catalog TEXT,
    is_summarized BOOLEAN NOT NULL DEFAULT FALSE,
    vec {{if isPostgres }} vector({{ .VectorSize }}) {{ else }} FLOAT[{{ .VectorSize }}] {{ end }} -- type description embedding
);

CREATE TABLE IF NOT EXISTS modules (
    name TEXT NOT NULL PRIMARY KEY,
	description TEXT NOT NULL DEFAULT '',
    long_description TEXT NOT NULL DEFAULT '',
    query_root TEXT REFERENCES types(name),
    mutation_root TEXT REFERENCES types(name),
    function_root TEXT REFERENCES types(name),
    mut_function_root TEXT REFERENCES types(name),
    is_summarized BOOLEAN NOT NULL DEFAULT FALSE,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    vec {{if isPostgres }} vector({{ .VectorSize }}) {{ else }} FLOAT[{{ .VectorSize }}] {{ end }} -- module description embedding
);

CREATE TABLE IF NOT EXISTS fields (
    type_name TEXT NOT NULL REFERENCES types(name),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    type TEXT NOT NULL REFERENCES types(name),
    hugr_type TEXT NOT NULL,
    catalog TEXT,
    is_indexed BOOLEAN NOT NULL DEFAULT FALSE,
    is_list BOOLEAN NOT NULL DEFAULT FALSE,
    is_non_null BOOLEAN NOT NULL DEFAULT FALSE,
    is_primary_key BOOLEAN NOT NULL DEFAULT FALSE,
    references_type TEXT REFERENCES types(name),
    mcp_exclude BOOLEAN NOT NULL DEFAULT FALSE,
    is_summarized BOOLEAN NOT NULL DEFAULT FALSE,
    vec {{if isPostgres }} vector({{ .VectorSize }}) {{ else }} FLOAT[{{ .VectorSize }}] {{ end }}, -- field description embedding
    PRIMARY KEY (type_name, name)
);

CREATE TABLE IF NOT EXISTS arguments (
    type_name TEXT NOT NULL REFERENCES types(name),
    field_name TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    type TEXT NOT NULL REFERENCES types(name),
    is_list BOOLEAN NOT NULL DEFAULT FALSE,
    is_non_null BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (type_name, field_name, name),
    FOREIGN KEY (type_name, field_name) REFERENCES fields(type_name, name)
);

CREATE TABLE IF NOT EXISTS data_sources (
    name TEXT NOT NULL PRIMARY KEY,
    description TEXT NOT NULL,
    long_description TEXT NOT NULL,
    type TEXT NOT NULL,
    prefix TEXT NOT NULL,
    as_module BOOLEAN NOT NULL DEFAULT FALSE,
    read_only BOOLEAN NOT NULL DEFAULT FALSE,
    is_summarized BOOLEAN NOT NULL DEFAULT FALSE,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    vec {{if isPostgres }} vector({{ .VectorSize }}) {{ else }} FLOAT[{{ .VectorSize }}] {{ end }} -- data source long description embedding
);

CREATE TABLE IF NOT EXISTS data_objects (
    name TEXT NOT NULL REFERENCES types(name) PRIMARY KEY,
    filter_type_name TEXT REFERENCES types(name),
    args_type_name TEXT REFERENCES types(name)
);

CREATE TABLE IF NOT EXISTS data_object_queries (
    name TEXT NOT NULL,
    object_name TEXT NOT NULL REFERENCES types(name),
    query_root TEXT NOT NULL REFERENCES types(name),
    query_type TEXT NOT NULL,
    PRIMARY KEY (name, object_name)
);
