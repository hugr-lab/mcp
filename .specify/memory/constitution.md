<!--
Sync Impact Report:
- Version change: 1.0.0 → 2.0.0 (MAJOR: Added core authentication and schema indexing principles)
- Modified principles: 
  ✅ I. Schema-First Development - enhanced with dual-auth model
  ✅ IV. Lazy Stepwise Discovery - enhanced with vector search requirements
- Added sections: 
  ✅ VI. Dual Authentication Model
  ✅ VII. Schema Indexing & Enrichment
  ✅ VIII. Vector Search Foundation
  ✅ IX. Arrow-Native Data Transport
  ✅ Development Workflow - enhanced with build tags and dependencies
- Removed sections: N/A
- Templates requiring updates:
  ⚠️  .specify/templates/plan-template.md (add authentication section)
  ⚠️  .specify/templates/spec-template.md (add indexing requirements)
  ⚠️  .specify/templates/tasks-template.md (add schema sync tasks)
- Follow-up TODOs:
  - Update system_prompt.md with auth flow
  - Add indexer initialization to setup tasks
  - Document LLM summarization templates
-->

# Hugr MCP Service Constitution

## Core Principles

### I. Schema-First Development
The MCP service MUST expose Hugr's GraphQL schema through MCP tools and resources.
Every feature begins with schema introspection capabilities: types, fields, enums,
and discovery tools MUST be independently accessible and testable. Schema structure
determines tool boundaries - no organizational-only groupings.

**Dual Authentication Model**:
- **Schema Discovery**: Uses Secret key for unrestricted access to full schema metadata
- **Query Execution**: Forwards user's OIDC token to respect role-based field visibility and row-level filters
- Authentication managed by `pkg/auth.HugrTransport` (http.RoundTripper implementation)

**Schema Source**: Full schema obtained via `schema_summary` query:
```graphql
query { 
  function { 
    core {
      meta {
        schema_summary
      }
    }
  }
}
```

**Rationale**: Hugr's role-based filtering requires token forwarding for user queries,
but schema discovery needs unrestricted access. Dual auth ensures agents can discover
full capabilities while queries execute with proper authorization.

### II. Tool-Based Interface
Every capability MUST be exposed as an MCP tool or resource following text-in/JSON-out
protocol. Tools accept structured JSON inputs, return structured JSON outputs, errors
go to standard MCP error responses. Support both discovery tools (search, introspection)
and execution tools (GraphQL queries, function calls).

**Rationale**: MCP protocol standardization ensures compatibility across AI agents
(Claude, GPT, etc.) and enables composable, testable interactions.

### III. Test-First Development (NON-NEGOTIABLE)
TDD is mandatory: Tests written → User approved → Tests fail → Then implement.
Red-Green-Refactor cycle strictly enforced. No implementation without failing tests.
Contract tests for tool schemas, integration tests for Hugr interactions.

**Rationale**: Schema introspection and query generation are complex; TDD ensures
correctness and prevents regressions in agent-facing APIs.

### IV. Lazy Stepwise Discovery
Tools MUST support progressive schema exploration: broad discovery → refinement →
detailed introspection. Never assume fixed schema names. Vector search for semantic
field/module discovery MUST be available where embeddings exist. Filter and aggregate
early to limit data transfer.

**Discovery Flow**:
1. **High-level**: Data sources and root modules (via vector search or hierarchical listing)
   - Query: `{ __schema { queryType { name, fields { name, description } } } }`
   - Identify module namespaces and their purposes
   - Check `data_sources` table for source integration mode (`as_module` flag)
   
2. **Module hierarchy**: Navigate submodules recursively and understand capabilities
   - For each module: Determine which root types exist (query, mutation, function, mut_function)
   - Build path map with arbitrary depth: `module.sub1.sub2...subN.capability`
   - Handle data source prefixes in type names
   - Example paths:
     - Data: `core.mcp.types` (depth 2)
     - Function: `function.analytics.reports.sales.quarterly.calculate` (depth 5)
     - Source as module: `postgres_db.public.users` (source integrated as module)

3. **Mid-level**: Tables/views/functions within modules (with semantic ranking)
   - Query module's query type recursively to list available tables/views at any depth
   - Query module's function type recursively to list available functions
   - Use vector search: "Find tables related to user authentication in module X"
   - Apply prefix awareness: search for `pg_users` if prefix is "pg_"

4. **Detail-level**: Field-level introspection with nested types and arguments
   - Full type definition: fields, arguments, nested relationships
   - Function signatures: parameters, return types
   - Mutation availability: insert/update/delete operations
   - Source prefix resolution: map prefixed type names back to original table names

**Module Path Construction** (arbitrary depth):
MCP tools MUST construct GraphQL paths respecting module hierarchy with no depth limits:
```graphql
# Data query path: {root}.{module}.{sub1}...{subN}.{table}
query { 
  company { 
    department { 
      team { 
        project { 
          mcp_members { ... }  # Note: prefix "mcp_" from data source
        } 
      } 
    } 
  } 
}

# Function path: function.{module}.{sub1}...{subN}.{function_name}
query { 
  function { 
    analytics { 
      processing { 
        aggregation { 
          compute_metrics(...) 
        } 
      } 
    } 
  } 
}

# Data source as module: {source_module}.{schema/namespace}.{table}
query {
  analytics_db {      # Data source registered as module
    public {          # Database schema
      sales { ... }   # Table (no prefix needed - module isolation)
    }
  }
}

# Mixed sources in single module
query {
  reporting {
    pg_sales { ... }       # PostgreSQL table (prefix: "pg_")
    ddb_aggregates { ... } # DuckDB table (prefix: "ddb_")
    es_logs { ... }        # ElasticSearch index (prefix: "es_")
  }
}
```

**Module Type Resolution**:
- `modules.query_root` → GraphQL type containing data tables/views OR submodules
- `modules.mutation_root` → GraphQL type containing insert/update/delete mutations OR submodules
- `modules.function_root` → GraphQL type containing read-only functions OR submodules
- `modules.mut_function_root` → GraphQL type containing state-changing functions OR submodules
- Each root type is a GraphQL OBJECT with fields that may be:
  - **Leaf capabilities**: Tables, views, functions (queryable entities)
  - **Intermediate nodes**: Submodules (nested OBJECT types) for arbitrary depth
- Recursive traversal required to reach leaf capabilities in deep hierarchies

**Data Source Prefix Resolution**:
- When discovering types, check `types.catalog` field for data source association
- Lookup `data_sources.prefix` to understand type naming
- MCP tools should expose both prefixed name (GraphQL) and original name (documentation)
- Example: Type `mcp_types` → catalog: "mcp_indexer", prefix: "mcp_", original: "types"

**Rationale**: Hugr schemas are dynamic and role-dependent. Agents need discovery
patterns to navigate unfamiliar or evolving data structures efficiently. Module
hierarchy requires path-aware query construction. Vector search enables semantic
matching ("customer data", "geospatial queries") within module context.

### V. Observability & Error Clarity
Structured logging required for all tool invocations, schema queries, and Hugr
interactions. Errors MUST include actionable context (missing fields, invalid filters,
auth failures). Support debug mode for query introspection.

**Rationale**: Text I/O ensures debuggability. Agents and users need clear feedback
when schema access fails or queries are malformed.

### VI. Dual Authentication Model (NON-NEGOTIABLE)
Service MUST support two authentication contexts simultaneously:

1. **Service-Level Auth (Secret Key)**:
   - Purpose: Schema discovery, metadata queries, indexing operations
   - Scope: Unrestricted access to full schema structure
   - Used by: `schema_summary` queries, indexer initialization
   - Configuration: Environment variable `HUGR_SECRET_KEY`

2. **User-Level Auth (OIDC Token)**:
   - Purpose: Query execution on behalf of users
   - Scope: Role-based field visibility, mandatory row-level filters
   - Used by: All data queries via MCP tools
   - Flow: Token extracted from MCP request → forwarded via `HugrTransport`

**Implementation**: `pkg/auth.HugrTransport` implements `http.RoundTripper`:
```go
type HugrTransport struct {
    BaseTransport http.RoundTripper
    SecretKey     string          // For schema queries
    GetUserToken  func() string   // Extracts token from request context
}
```

**Rationale**: Hugr's role-filtering (field visibility, mandatory filters) requires
forwarding user tokens, but service needs unrestricted schema access for indexing
and discovery. Dual auth prevents privilege escalation while enabling full metadata.

### VII. Schema Indexing & Enrichment (NON-NEGOTIABLE)
Service MUST maintain indexed schema metadata for vector search and semantic discovery.
**All database operations execute through Hugr GraphQL API after registration.**

**Initialization Flow**:

1. **Database Setup** (pkg/indexer):
   - On startup, check if indexer database exists in Hugr (query data sources via GraphQL)
   - If absent: 
     - Create database with schema (PostgreSQL with pgvector OR DuckDB with vss)
     - Register as Hugr data source using Hugr schema definition
     - **After registration: All subsequent access via Hugr GraphQL API only**
   - Supported databases: PostgreSQL (pgvector extension), DuckDB (vss extension)
   - Schema versioning via `version_info` table

2. **Dual Schema Extraction** (GraphQL queries executed simultaneously):
   
   **Query A - Introspection** (core type system):
   ```graphql
   query schema {
     __schema {
       description
       queryType { ...type_info }
       mutationType { ...type_info }
       types { ...type_info }
     }
   }
   fragment type_info on __Type {
     name, description, kind, hugr_type, module, catalog
     enumValues { name, description }
     inputFields { name, description, type {...} }
     fields { name, description, hugr_type, catalog, args {...}, type {...} }
   }
   ```
   
   **Query B - Enhanced Metadata** (executed in parallel):
   ```graphql
   query { 
     function { 
       core { 
         meta { 
           schema_summary 
         } 
       } 
     } 
   }
   ```
   Returns: SchemaInfo → DataSourceInfo[], ModuleInfo, DataObjectInfo[], FunctionInfo[], FieldInfo[]

3. **Data Population via Hugr GraphQL**:
   - Parse introspection + schema_summary results
   - Insert into indexer database **using Hugr mutations**:
     - **Single-record inserts** (can batch multiple mutations):
       ```graphql
       mutation ($input: mcp_types_mut_input_data!, $summary: String!) {
         core {
           mcp {
             insert_types(data: $input, summary: $summary) {
               name
             }
           }
         }
       }
       ```
     - `$input`: Single record data (not array)
     - `$summary`: Description for embedding generation
     - Batch multiple insert mutations in single request
   
   - **Bulk updates** (filter-based, multiple records):
     ```graphql
     mutation ($filter: mcp_types_filter!, $data: mcp_types_mut_data!, $summary: String!) {
       core {
         mcp {
           update_types(
             filter: $filter
             data: $data
             summary: $summary
           ) {
             success
           }
         }
       }
     }
     ```
   
   - **Primary key updates** (for vector/embedding updates):
     ```graphql
     mutation ($name: String!, $type_name: String!, $input: mcp_fields_mut_data!, $summary: String!) {
       core {
         mcp {
           update_fields(
             filter: { name: { eq: $name }, type_name: { eq: $type_name } }
             data: $input
             summary: $summary
           ) {
             success
           }
         }
       }
     }
     ```
     - Vector updates MUST use primary key filters (one embedding per mutation)
     - Hugr computes embedding from `$summary` parameter
   
   - **Incremental schema updates**:
     Use specialized functions for targeted updates when Hugr schema changes:
     - `LoadDataObject(name)`: Update single data object without full re-index
     - `LoadFunction(module, name)`: Update single function definition
     - `LoadModule(name)`: Update module metadata
     - `LoadDataSource(name)`: Update data source information
     - Avoids full schema re-extraction and re-summarization
   
   - Populate tables via Hugr: `types`, `fields`, `arguments`, `modules`, `data_sources`, `data_objects`, `data_object_queries`
   - No direct database access after registration

4. **Vector Embeddings Generation**:
   - Generate embeddings for searchable fields:
     - `data_sources.vec`: From long_description
     - `modules.vec`: From long_description  
     - `types.vec`: From long_description
     - `fields.vec`: From description
   - Update via Hugr GraphQL mutations:
     ```graphql
     mutation update_embeddings($name: String!, $vec: [Float!]!) {
       core {
         mcp {
           update_types(
             filter: {name: {eq: $name}}
             set: {vec: $vec}
           ) {
             affected_rows
           }
         }
       }
     }
     ```

5. **LLM Summarization** (pkg/summary - OPTIONAL):
   - Generate business-friendly `long_description` via LLM templates
   - Supported providers: OpenAI, Claude, Custom (LM Studio compatible)
   - Update records via Hugr mutations:
     - Use primary key filters for vector updates
     - Pass `summary: $long_description` to trigger embedding generation
     - Set `is_summarized: true` in mutation data
   - Managed by `pkg/pool` for LLM connections
   - Targets: data_sources, modules, types, fields

6. **Incremental Updates** (pkg/indexer functions):
   When Hugr schema changes, use targeted update functions instead of full re-index:
   - `LoadDataObject(name string)`: Update single table/view metadata
   - `LoadFunction(module, name string)`: Update function definition
   - `LoadModule(name string)`: Update module hierarchy and root types
   - `LoadDataSource(name string)`: Update external data source info
   
   **Benefits**:
   - Avoid full `__schema` introspection query
   - Preserve existing summaries (no re-summarization)
   - Faster updates for incremental schema evolution
   - Reduces LLM API calls and embedding computations
   
   **Implementation**:
   - Query specific schema elements via GraphQL
   - Execute targeted mutations with primary key filters
   - Update `long_description` only if changed
   - Re-generate embeddings only when summary modified

**Database Schema**:
Core tables follow Hugr schema definition (see attached schema template):
- `types`: GraphQL types with kind, module, catalog metadata
  - `catalog`: References `data_sources.name` to identify source and its prefix
- `modules`: Hierarchical organization with unlimited nesting depth
  - `query_root`: Type containing data query fields (tables/views) OR submodule fields
  - `mutation_root`: Type containing mutation fields (insert/update/delete) OR submodule fields
  - `function_root`: Type containing read-only function fields OR submodule fields
  - `mut_function_root`: Type containing mutation function fields OR submodule fields
  - Root types are GraphQL OBJECT types - fields may be capabilities or nested modules
  - No explicit parent-child table - hierarchy inferred from type field structure
- `fields`: Type fields with arguments, references, and flags
  - Module root types contain fields that are either:
    - Submodules: Field type is another module's root type (intermediate node)
    - Capabilities: Field type is data table/view/function (leaf node)
  - Recursive structure enables arbitrary depth
- `arguments`: Field arguments with type information
- `data_sources`: External data systems with integration configuration
  - `prefix`: String prepended to all GraphQL types/inputs from this source
  - `as_module`: Boolean - if true, source becomes standalone module; if false, types scattered across modules
  - `type`: Source type (PostgreSQL, DuckDB, ElasticSearch, etc.)
- `data_objects`: Queryable entities with filter/args types
  - Type names include source prefix
- `data_object_queries`: Available query operations with query_root reference
  - Links to module's query type for path construction
- Views: `module_intro` for flat module exploration across all four root types

**Module Hierarchy Representation** (unlimited depth):
```
Module "company" (root level)
├─ query_root: "company_query" (OBJECT type)
│  ├─ Field "department" → Type "company_dept_query" (submodule - intermediate)
│  │  ├─ Field "team" → Type "company_dept_team_query" (submodule - intermediate)
│  │  │  ├─ Field "project" → Type "company_dept_team_proj_query" (submodule - intermediate)
│  │  │  │  ├─ Field "mcp_members" → Type "[mcp_member]" (table - leaf, prefix "mcp_")
│  │  │  │  └─ Field "pg_tasks" → Type "[pg_task]" (table - leaf, prefix "pg_")
│  │  │  └─ Field "pg_teams" → Type "[pg_team]" (table - leaf)
│  │  └─ Field "ddb_metrics" → Type "[ddb_metric]" (table - leaf, prefix "ddb_")
│  └─ Field "analytics" → Type "company_analytics_query" (submodule)
│     └─ ... (more nesting)
│
├─ mutation_root: "company_mutation" (OBJECT type)
│  └─ Field "department" → Type "company_dept_mutation" (submodule)
│     └─ Field "team" → ...
│        └─ Field "insert_mcp_member" → Mutation function (leaf)
│
├─ function_root: "company_function" (OBJECT type)
│  └─ Field "analytics" → Type "company_analytics_function" (submodule)
│     └─ Field "reports" → ...
│        └─ Field "generate" → Function (leaf)
│
└─ mut_function_root: "company_mut_function" (OBJECT type)
   └─ Field "processing" → ...

Data Source as Module:
Module "postgres_main" (registered with as_module: true, prefix: "")
├─ query_root: "postgres_main_query"
│  ├─ Field "public" → Type "postgres_main_public_query" (schema namespace)
│  │  ├─ Field "users" → Type "[user]" (table - no prefix needed)
│  │  └─ Field "orders" → Type "[order]" (table)
│  └─ Field "analytics" → Type "postgres_main_analytics_query" (schema namespace)
│     └─ Field "sales_summary" → Type "[sales_summary]" (view)
```

**Path Construction Logic**:
MCP tools resolve paths by:
1. Identify target capability type (query/mutation/function/mut_function)
2. Start from appropriate root: `query`, `mutation`, or `function` (for both function types)
3. Traverse module hierarchy recursively: `root.module.sub1.sub2...subN.capability`
4. Use `modules` table to map module → root_type → fields → submodules OR capabilities
5. Terminate when field type is not another module root (reached leaf capability)
6. Apply data source prefix when constructing type/mutation names

**Indexer Schema Registration**:
Database schema uses Hugr schema language:
- `@table`, `@view` directives for table/view mapping
- `@embeddings(model, vector, distance)` enables vector search per table
  - Exposes `_distance_to_query(query: String!)` field for semantic ranking
  - Automatic score field generation for similarity results
- `@field_references` establishes queryable relationships
- `@pk` marks primary keys
- `Vector` type with `@dim(len: N)` for embedding dimensions

**Access Pattern**:
After registration, indexer database becomes a Hugr module (`core.mcp`). All operations
use GraphQL queries/mutations with `core.mcp` prefix. No direct SQL access permitted.

**Data Structures**: Use Go types from specification (SchemaInfo, DataObjectInfo,
FieldInfo, etc.) for consistent schema representation. Match field names to database
columns for GraphQL query generation.

**Rationale**: Hugr schemas are complex and domain-specific. Dual-query extraction
(introspection + schema_summary) executed simultaneously captures both GraphQL type
system and Hugr-specific metadata efficiently. Hugr-mediated access ensures role-based
filtering applies to indexer queries. Vector search on enriched descriptions enables
semantic discovery ("find customer revenue data") vs exact name matching.

### VIII. Vector Search Foundation
Vector search MUST be available for semantic schema discovery. Agents query by
intent ("geospatial analysis", "time-series metrics") rather than exact names.
**All search operations execute through Hugr GraphQL API using `_distance_to_query` field.**

**Search Capabilities**:
- **Data Source search**: Find external systems by description (e.g., "production analytics database")
- **Module search**: Find modules by business domain (e.g., "customer management", "billing")
- **Type search**: Locate types by purpose (e.g., "user profile", "transaction record")
- **Field search**: Discover fields by semantic meaning with relevance scores (e.g., "shipping address", "total revenue")

**Search Implementation via Hugr GraphQL**:
Hugr's `@embeddings` directive automatically exposes `_distance_to_query(query: String!)` field
for vector similarity. Example query for semantic type search with nested field ranking:

```graphql
query ranked_data_objects(
  $filter: mcp_types_filter!,
  $query: String!,
  $fieldsQuery: String!,
  $ttl: Int!
) {
  core {
    mcp {
      types(
        filter: $filter
        order_by: [{field: "score"}]
      ) @cache(ttl: $ttl) {
        name
        module
        description
        long_description
        type: hugr_type
        score: _distance_to_query(query: $query)
        fields(
          nested_order_by: [{field: "score"}]
        ) {
          name
          description
          type
          hugr_type
          mcp_exclude
          is_primary_key
          is_list
          is_non_null
          score: _distance_to_query(query: $fieldsQuery)
        }
        data_object {
          args_type_name
          queries {
            name
            query_root
            query_type
            field {
              description
              mcp_exclude
              type
              is_list
              arguments {
                name
                description
                type
                is_non_null
                is_list
              }
            }
          }
        }
      }
    }
  }
}
```

**Search Features**:
- **Semantic ranking**: `_distance_to_query(query: $query)` generates similarity score
- **Nested ordering**: Fields ranked by relevance within parent types
- **Multi-level search**: Separate queries for types ($query) and fields ($fieldsQuery)
- **Caching**: `@cache(ttl: $ttl)` for frequently searched patterns
- **Filtering**: Combine semantic search with structural filters (e.g., `{module: {eq: "core"}}`)

**Embedding Strategy**:
- Hugr handles embedding generation when `@embeddings` directive present
- Model configured via schema template: `@embeddings(model: "{{ .EmbeddingModel }}", ...)`
- Distance metric: Cosine (specified in directive)
- Dimension configured: `Vector @dim(len: {{ .VectorSize }})`
- Service only provides query strings; Hugr computes embeddings and similarities

**MCP Tool Integration**:
Tools MUST support parameters for GraphQL-based searches:
- `query`: Semantic search string for primary entities
- `fields_query`: Separate search string for nested fields (optional)
- `top_k`: Maximum results via GraphQL `limit` (default 5, max 50)
- `min_score`: Minimum similarity threshold via GraphQL `filter` on score field
- `ttl`: Cache duration in seconds for repeated searches
- Return results sorted by `score` field (highest first)

**View Support**:
`module_intro` view provides flattened module exploration for semantic search across
module capabilities (queries, mutations, functions) without complex joins. Also supports
`_distance_to_query` for module-level semantic discovery.

**No Direct Database Access**:
Service MUST NOT execute raw SQL for vector searches. All operations via Hugr GraphQL:
- Schema registration enables `@embeddings` directive processing
- Hugr manages embedding models, storage, and similarity computations
- Service constructs GraphQL queries with `_distance_to_query` fields

**Implementation**: `pkg/indexer` handles:
- GraphQL query construction with `_distance_to_query` fields
- Result parsing and score interpretation
- Filter composition (semantic + structural)
- Cache management via `@cache` directive

**Rationale**: Dynamic schemas and role-based visibility make exact name queries
impractical. Semantic search adapts to terminology variations. Hugr-mediated vector
search ensures consistent embedding models and respects role-based field filtering.
View-based search improves performance for common module discovery patterns.

### IX. Arrow-Native Data Transport
Service MUST support Apache Arrow IPC for efficient data transfer from Hugr.

**Build Configuration**:
```jsonc
// .vscode/settings.json or build flags
"go.buildFlags": ["-tags=duckdb_arrow"],
"go.testEnvVars": {
    "CGO_CFLAGS": "-O1 -g"
},
"go.testTags": "duckdb_arrow"
```

**Rationale**: Hugr uses Arrow IPC for zero-copy data transport. Native Arrow support
enables efficient handling of large result sets and nested data structures.

## Development Workflow

### Development Workflow

### Code Organization
- **Single Go module**: Root `cmd/server/` for MCP server, `pkg/` for tools/services
- **Package structure**:
  - `pkg/tools/`: MCP tool implementations (schema, discovery, execution)
    - Must handle module path construction for GraphQL queries with unlimited depth
    - Navigate four root types: query, mutation, function, mut_function
    - Resolve data source prefixes when constructing type/mutation names
    - Recursive path traversal for arbitrary nesting levels
  - `pkg/auth/`: HugrTransport for dual authentication
  - `pkg/indexer/`: Schema indexing via Hugr GraphQL
    - GraphQL query/mutation construction with module paths (unlimited depth)
    - Module hierarchy traversal and path resolution (recursive)
    - Data source prefix tracking and resolution
    - Result parsing and caching
    - Incremental update functions: `LoadDataObject()`, `LoadFunction()`, `LoadModule()`, `LoadDataSource()`
  - `pkg/summary/`: LLM-based description enrichment (optional)
  - `pkg/pool/`: LLM client pool management
- **Testing**: 
  - `tests/contract/`: MCP tool schemas
  - `tests/integration/`: Hugr API interactions with auth mocking
  - `tests/indexer/`: Vector search accuracy and GraphQL query correctness
  - `tests/paths/`: Module path construction and resolution correctness
  - Arbitrary depth traversal (5+ levels)
  - Data source prefix resolution
  - Mixed-source module paths
- **Documentation**: `pkg/service/system_prompt.md` contains agent guidance
  - Must document module hierarchy navigation
  - Examples of path construction for all four root types

### External Dependencies
**Core**:
- `github.com/hugr-lab/query-engine`: Hugr GraphQL client
  - Official documentation: https://hugr-lab.github.io
  - Source repositories (for implementation details):
    - Core: https://github.com/hugr-lab/hugr
    - Query Engine: https://github.com/hugr-lab/query-engine
    - Docker/Config: https://github.com/hugr-lab/docker
  - Consult docs first; check source code for complex/undocumented behavior
- `github.com/mark3labs/mcp-go/mcp`: MCP server framework
- `github.com/tmc/langchaingo`: LLM integrations (OpenAI, Claude, custom)

**Data & Search**:
- PostgreSQL driver + pgvector extension (optional)
- DuckDB with vss extension (optional)
- Apache Arrow Go libraries

**Development**:
- Build tags: `duckdb_arrow` for Arrow support
- CGO required for DuckDB (set `CGO_CFLAGS=-O1 -g` for debugging)

**Reference Materials** (in order of preference):
1. Hugr Documentation: https://hugr-lab.github.io/docs/
   - GraphQL API, Schema Definition, Authentication
2. Hugr Source Code (when docs insufficient):
   - Core engine tests: https://github.com/hugr-lab/hugr/tree/main/tests
   - Query engine examples: https://github.com/hugr-lab/query-engine/tree/main/examples
   - Deployment configs: https://github.com/hugr-lab/docker/tree/main/examples
3. Test files in repositories often document edge cases better than formal docs

### Quality Gates
- All PRs MUST pass contract tests (MCP tool schemas)
- Integration tests MUST verify Hugr GraphQL query correctness with auth scenarios
- Indexer tests MUST validate vector search relevance scores
- **Path resolution tests** MUST verify correct module hierarchy navigation
  - Test all four root types (query, mutation, function, mut_function)
  - Test deeply nested submodule paths with arbitrary depth (e.g., `a.b.c.d.e.f.capability`)
  - Test path construction from `modules` table metadata
  - Test data source prefix resolution in type names
  - Test both integration modes: `as_module: true` vs `as_module: false`
  - Test mixed-source modules (multiple prefixed types in one module path)
- System prompt updates MUST align with new/changed tools
- No merge without test coverage for:
  - Dual auth scenarios (schema vs query execution)
  - Schema indexing and updates
  - Vector search accuracy
  - Module path construction and resolution (unlimited depth)
  - Data source prefix handling (both integration modes)

### Versioning & Breaking Changes
- MCP tool schema changes follow semantic versioning
- MAJOR: Remove tool, change auth model, incompatible schema format
- MINOR: Add tool, add optional parameters, extend output fields
- PATCH: Fix bugs, clarify errors, performance improvements
- Schema changes in Hugr MUST trigger:
  - Re-indexing validation
  - Tool compatibility review
  - System prompt updates

## Governance

### Amendment Process
1. Propose change via PR to `.specify/memory/constitution.md`
2. Update version per semantic rules (justify in PR description)
3. Validate dependent templates (plan, spec, tasks) for consistency
4. Update `.specify/templates/commands/*.md` if workflows change
5. Update `pkg/service/system_prompt.md` if agent guidance changes
6. Merge requires approval + all template sync checks passed

### Compliance Review
- Constitution supersedes all other practices
- All PRs/reviews MUST verify constitutional compliance
- Complexity deviations MUST be justified in `plan.md` Complexity Tracking
- Agent guidance (`system_prompt.md`) MUST reflect current tool capabilities
- Auth implementation MUST prevent privilege escalation (no user tokens for indexing)
- **Hugr behavior verification protocol**:
  1. Check documentation: https://hugr-lab.github.io
  2. If docs unclear/incomplete, examine source code:
     - Directive behavior → https://github.com/hugr-lab/hugr (compiler internals)
     - Query execution → https://github.com/hugr-lab/query-engine (resolver logic)
     - Auth/config → https://github.com/hugr-lab/docker (deployment examples)
  3. Review test files for edge case documentation
  4. Document findings in PR if behavior is undocumented
  5. Never implement based on assumptions; verify through code or tests

### Runtime Guidance
Use `pkg/service/system_prompt.md` for agent-specific development guidance and tool usage patterns.
Keep this file synchronized with MCP tool implementations and authentication flows.

## Technical Constraints Summary

### Authentication
- **Schema queries**: Secret key (unrestricted access to full schema)
- **User queries**: OIDC token forwarding (role-filtered field visibility and row filters)
- **Implementation**: `pkg/auth.HugrTransport`
- **Environment**: `HUGR_SECRET_KEY` for service-level auth

### Schema Indexing
- **Database**: PostgreSQL (pgvector) or DuckDB (vss extension)
- **Initialization**: 
  - Execute simultaneously:
    - Query A: GraphQL `__schema` introspection → Core types/fields
    - Query B: `schema_summary` → Enhanced metadata (data sources, modules, objects)
  - Create database if missing → Register as Hugr data source
  - **After registration: All access via Hugr GraphQL API only (no direct SQL)**
- **Schema Definition**: Hugr schema language with directives
  - `@table`, `@view`: Table/view mapping
  - `@embeddings(model, vector, distance)`: Vector search config
    - Auto-generates `_distance_to_query(query: String!)` field for semantic ranking
  - `@field_references`: Queryable relationships
  - `@pk`: Primary key markers
  - `Vector @dim(len: N)`: Embedding columns
- **Data Population**: Via Hugr GraphQL mutations:
  - **Insert**: Single record per mutation, batch multiple in request
    ```graphql
    mutation ($input: mcp_types_mut_input_data!, $summary: String!) {
      core { mcp { insert_types(data: $input, summary: $summary) { name } } }
    }
    ```
  - **Update (bulk)**: Filter-based for multiple records
  - **Update (vector)**: Primary key filter only (one embedding per mutation)
    ```graphql
    mutation ($name: String!, $data: mcp_types_mut_data!, $summary: String!) {
      core { mcp { update_types(filter: {name: {eq: $name}}, data: $data, summary: $summary) { success } } }
    }
    ```
  - `$summary` parameter triggers Hugr's embedding generation
- **Update strategy**: 
  - On-demand re-indexing via MCP tool (triggers Hugr mutations)
  - Scheduled background sync (optional, via GraphQL queries)
  - Schema version tracking via `version_info` table
- **Package**: `pkg/indexer`

### Database Schema Tables
Core tables (see attached schema):
- `types`: GraphQL types (name, kind, module, catalog, vec)
- `modules`: Module hierarchy (name, root type refs, vec)
- `fields`: Type fields (type_name, name, type, hugr_type, vec)
- `arguments`: Field arguments (type_name, field_name, name, type)
- `data_sources`: External systems (name, type, prefix, vec)
- `data_objects`: Queryable entities (name, filter_type, args_type)
- `data_object_queries`: Query operations (name, object_name, query_type)
- `version_info`: Schema migration tracking

Views:
- `module_intro`: Flattened module capabilities for search

**Access Pattern**: Indexer database registered as `core.mcp` module in Hugr.
All queries/mutations use GraphQL with `core.mcp` prefix. No direct database drivers.

### LLM Summarization (Optional)
- **Providers**: OpenAI, Claude, Custom (LM Studio compatible)
- **Purpose**: Generate `long_description` from `description`
- **Targets**: data_sources, modules, types, fields
- **Flags**: `is_summarized` boolean per record
- **Packages**: `pkg/summary` (templates), `pkg/pool` (LLM connections)

### Data Transport
- **Arrow IPC support**: Mandatory for Hugr result handling
- **Build tags**: `duckdb_arrow` (required)
- **CGO flags**: `-O1 -g` for debugging DuckDB
- **Rationale**: Hugr uses Arrow for zero-copy data transfer

### Vector Search
- **Semantic discovery**: Modules, types, fields, data sources
- **Implementation**: Hugr GraphQL with `_distance_to_query(query: String!)` field
  - Auto-generated by `@embeddings` directive
  - Service provides query strings; Hugr computes embeddings and similarities
- **Distance metric**: Cosine (configured in `@embeddings` directive)
- **Relevance scoring**: Score field in query results (0.0-1.0 scale, higher = more similar)
- **Query parameters**: 
  - `query`: Semantic search string for primary entities
  - `fields_query`: Separate search for nested fields (optional)
  - `top_k`: Max results via GraphQL `limit` (default 5, max 50)
  - `min_score`: Threshold via GraphQL `filter` on score field (default 0.3)
  - `ttl`: Cache duration in seconds (via `@cache` directive)
- **Features**:
  - Nested ordering: `nested_order_by: [{field: "score"}]`
  - Multi-level search: Different queries for types vs fields
  - Structural filters: Combine semantic with exact matches
- **No Direct Database Access**: All searches via Hugr GraphQL queries
- **Package**: `pkg/indexer` (constructs GraphQL queries, parses results)

### GraphQL Queries
**Introspection** (initial load - Query A):
```graphql
query schema {
  __schema {
    description
    queryType { ...type_info }
    mutationType { ...type_info }
    types { ...type_info }
  }
}
fragment type_info on __Type {
  name, description, kind, hugr_type, module, catalog
  enumValues { name, description }
  inputFields { ... }
  fields { name, description, hugr_type, args {...}, type {...} }
}
```

**Schema Summary** (enhanced metadata - Query B, executed simultaneously with Query A):
```graphql
query { 
  function { 
    core { 
      meta { 
        schema_summary 
      } 
    } 
  } 
}
```
Returns: SchemaInfo → DataSourceInfo[], ModuleInfo, DataObjectInfo[], FunctionInfo[], FieldInfo[]

**Vector Search Example** (semantic type discovery with nested field ranking):
```graphql
query ranked_data_objects(
  $filter: mcp_types_filter!,
  $query: String!,
  $fieldsQuery: String!,
  $ttl: Int!
) {
  # Path: query root → core module → mcp submodule → types table
  core {
    mcp {
      types(
        filter: $filter
        order_by: [{field: "score"}]
      ) @cache(ttl: $ttl) {
        name
        module
        score: _distance_to_query(query: $query)
        fields(nested_order_by: [{field: "score"}]) {
          name
          description
          score: _distance_to_query(query: $fieldsQuery)
        }
      }
    }
  }
}
```

**Module-aware query construction** (arbitrary depth with prefixes):
```graphql
# Query data in deeply nested module with prefixed types
query {
  # Path: company.department.team.project.mcp_members
  # Note: "mcp_" prefix from data source
  company {
    department {
      team {
        project {
          mcp_members(filter: {active: {eq: true}}) {
            id
            name
            role
          }
        }
      }
    }
  }
}

# Query data source integrated as module (no prefix needed)
query {
  # Path: postgres_main.public.users
  # Data source as module - natural schema namespaces
  postgres_main {
    public {
      users(filter: {email: {like: "%@company.com"}}) {
        id
        email
        created_at
      }
    }
  }
}

# Mixed sources in single module path
query {
  # Module combining multiple prefixed sources
  reporting {
    pg_sales(filter: {year: {eq: 2025}}) {     # PostgreSQL (prefix: "pg_")
      total_revenue
    }
    ddb_aggregates {                            # DuckDB (prefix: "ddb_")
      monthly_summary
    }
    es_logs(filter: {level: {eq: "ERROR"}}) {  # ElasticSearch (prefix: "es_")
      timestamp
      message
    }
  }
}

# Call function in deeply nested module
query {
  # Path: function.analytics.processing.aggregation.financial.compute_metrics
  function {
    analytics {
      processing {
        aggregation {
          financial {
            compute_metrics(
              period: "Q1_2025",
              currency: "USD"
            ) {
              result
              metadata
            }
          }
        }
      }
    }
  }
}

# Mutation in nested module with prefix
mutation {
  # Path: company.department.team.insert_mcp_member
  # Note: mutation name includes prefix
  company {
    department {
      team {
        insert_mcp_member(
          data: {
            name: "John Doe",
            role: "Developer"
          }
          summary: "New team member"
        ) {
          id
          name
        }
      }
    }
  }
}

# Update with source prefix in type filter
mutation {
  # Prefix in filter type: mcp_members_filter
  company {
    department {
      team {
        update_mcp_members(
          filter: {id: {eq: "member_123"}}
          data: {role: "Senior Developer"}
          summary: "Updated member role"
        ) {
          success
        }
      }
    }
  }
}
```

**Data Mutations** (populate schema data after registration):

**Single-record insert** (can batch multiple):
```graphql
mutation ($input: mcp_types_mut_input_data!, $summary: String!) {
  core {
    mcp {
      insert_types(data: $input, summary: $summary) {
        name
      }
    }
  }
}

mutation ($input: mcp_fields_mut_input_data!, $summary: String!) {
  core {
    mcp {
      insert_fields(data: $input, summary: $summary) {
        name
      }
    }
  }
}
```

**Bulk update** (filter-based):
```graphql
mutation ($filter: mcp_types_filter!, $data: mcp_types_mut_data!, $summary: String!) {
  core {
    mcp {
      update_types(
        filter: $filter
        data: $data
        summary: $summary
      ) {
        success
      }
    }
  }
}
```

**Primary key update** (for vector/embedding changes):
```graphql
mutation ($name: String!, $type_name: String!, $input: mcp_fields_mut_data!, $summary: String!) {
  core {
    mcp {
      update_fields(
        filter: { name: { eq: $name }, type_name: { eq: $type_name } }
        data: $input
        summary: $summary
      ) {
        success
      }
    }
  }
}
```

**Mutation Patterns**:
- **Insert**: Single record per mutation, batch multiple mutations in request
- **Update**: Filter-based for bulk, primary key filter for vector updates
- **Summary parameter**: Triggers embedding generation by Hugr
- **Vector updates**: MUST use PK filters (one embedding computation per mutation)

**Incremental Update Functions** (pkg/indexer):
- `LoadDataObject(name)`: Update single table/view without full re-index
- `LoadFunction(module, name)`: Update function definition
- `LoadModule(name)`: Update module metadata
- `LoadDataSource(name)`: Update data source info

**All indexer operations use Hugr GraphQL after database registration.**

---

**Version**: 2.0.0 | **Ratified**: 2025-10-06 | **Last Amended**: 2025-10-06

**Migration Notes from v1.0.0**:
- New authentication model requires environment variable `HUGR_SECRET_KEY`, if it was setup in the hugr deployment.
- Schema indexing database must be provisioned on first run
- Build configuration requires `duckdb_arrow` tag for Arrow support
- System prompt must document dual auth flow for agents
- **Hugr Reference Strategy**:
  - Start with documentation: https://hugr-lab.github.io
  - For complex/undocumented features, consult source code:
    - Schema directives & compiler: https://github.com/hugr-lab/hugr
    - GraphQL execution & client: https://github.com/hugr-lab/query-engine
    - Deployment & configuration: https://github.com/hugr-lab/docker
  - Test files often provide better edge case coverage than docs
  - Document any undocumented behavior discovered in implementation notes