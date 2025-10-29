# mcp
The Hugr mcp service

## Quick Start

### Local Development Environment (LDE)

```bash
# Start the LDE (from project root)
./lde.sh start

# Stop the LDE
./lde.sh stop

# Full cleanup
./lde.sh cleanup --force
```

See [lde.md](./lde.md) for complete LDE documentation.


Plan
1. Add new hugr scalar type vector (duckdb and postgres) `done`
2. Add to the hugr support for vector search (duckdb and postgres), if table contains vector fields accepts order by clause by vector similarity(define on field which type of distance metric to use) `done`
3. Extend indexer schema to include all tables `done`
4. Extend introspection types to return hugr types
5. Functions to define hugr type based on type directives: for types (table, view, module, table_agg, table_bucket_agg, filter_input, _h3_data, _join, _spatial, _h3), for fields (submodules, query, agg query, function, bucket_agg_query, h3, _spatial, _join, jq, _h3_data), for arguments (data object: filter, limit, offset, orderby, distinct, field name, field names, args, nested orderby, nested limit, nested ofset, vector search)
6. Fill data in data base for data sources, modules, tables, views, functions, fields, arguments `done`
7. Add summarization for modules, tables, views, functions `done`
8. Add feature for type and introspection attributes to exclude some fields from the schema (special directive @mcp_ignore) for tables and fields (using in extensions) to exclude fields that are should not be exposed to the user
9. Add feature to pass through the input values to the insert and update field default expressions that are not present in the schema as `@field_source(field: "-")` directive, this field should be excluded as a field in the introspection, but should be available in the inputs.
10. Add @embed directive to the fields descriptions to accept embedding model name to the field explanation
11. Vectorize fields descriptions (embedding generation)
12. Add vector search to find relevant fields for user query
13. Setup shell scripts and docker compose to run Developer and Testing environment with hugr required services (keycloak, postgres, redis, minio, etc. Also add synthea and open payments examples data to the hugr) done

--NEXT-- 
14. Split introspection and mcp data source queries to use different Hugr Clients (with different authentication and access rights). Hugr client that is used to access mcp tables and to perform type introspection queries.
15. Add CLI to create database, fill mcp db, summarize and attach db to the hugr if needed (to fill and summarize schema with out attaching the MCP db to the hugr).
16. Implement partial loading of the schema (load only new and changed modules, tables, views, functions, fields, arguments)
17. Implement partial summarization.
18. Implement http endpoints to manage the mcp service (partial loading and summarization)
19. Register mcp http service in the hugr to expose endpoints as GraphQL mutations
20. Implement the MCP client with OIDC authentication to access the MCP tools from the agents and UI

The rollout steps:
1. Using the CLI create DB and fill the schema
2. Using the CLI summarize the schema
3. Run service (automatically attach db and service endpoints to the hugr)
4. Using the MCP Client attach the MCP tools to the agent/UI/etc.


--NEXT--
21. Add MCP service to the hugr examples environment
22. Add MCP service to the helm chart to deploy with hugr
23. Add n8n to the LDE
24. Create example agents in n8n to use MCP tools
25. Add n8n integration to the MCP server (client with OIDC auth) to expose the agent mcp tools
26. Add simple UI to use MCP tools (chat and text outputs with incremental process status)
27. Implement resource to get general schema description (with modules and data sources)
28. Implement a logging mechanism to track user interactions and system performance
29. Implement a conversation artifacts management tool and artifacts storage
30. Implement a saved queries, to run saved queries with parameters (including vector search to find relevant queries)
31. Implement a tool to save query results as artifacts (tables, charts, maps) - session resources (also add session management to the client and server)
32. Add custom instructions support to the system prompt and summarization prompts to improve the quality of the generated queries and answers
--UI--
36. Implement a UI for chat interactions
37. Implement a UI for visualizations (tables, charts, maps)

