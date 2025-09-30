# mcp
The Hugr mcp service


Plan
1. Add new hugr scalar type vector (duckdb and postgres) `done`
2. Add to the hugr support for vector search (duckdb and postgres), if table contains vector fields accepts order by clause by vector similarity(define on field which type of distance metric to use) `done`
3. Extend indexer schema to include all tables `done`
4. Extend introspection types to return hugr types
5. Functions to define hugr type based on type directives: for types (table, view, module, table_agg, table_bucket_agg, filter_input, _h3_data, _join, _spatial, _h3), for fields (submodules, query, agg query, function, bucket_agg_query, h3, _spatial, _join, jq, _h3_data), for arguments (data object: filter, limit, offset, orderby, distinct, field name, field names, args, nested orderby, nested limit, nested ofset, vector search)
6. Fill data in data base for data sources, modules, tables, views, functions, fields, arguments `done`
7. Add summarization for modules, tables, views, functions `done`
--NEXT--
8. Add feature for type and introspection attributes to exclude some fields from the schema (special directive @mcp_ignore) for tables and fields (using in extensions) to exclude fields that are should not be exposed to the user
9. Add feature to pass through the input values to the insert and update field default expressions that are not present in the schema as `@field_source(field: "-")` directive, this field should be excluded as a field in the introspection, but should be available in the inputs.
10. Add @embed directive to the fields descriptions to accept embedding model name to the field explanation
11. Vectorize fields descriptions (embedding generation)
12. Add vector search to find relevant fields for user query
13. Implement resource to get general schema description (with modules and data sources)
14. Implement tool to type introspection (include vector search for the fields)
15. Implement tool to get field values in the table, view
16. Implement tool to run GraphQL queries
17. Implement tool to run functions
18. Implement Agents that will interact with the schema and perform actions based on user queries
19. Implement a logging mechanism to track user interactions and system performance
20. Implement a conversation artifacts management tool and artifacts storage
21. Implement a saved queries, to run saved queries with parameters (including vector search to find relevant queries)
22. Implement a oidc authentication (with hugr and keycloak to pass through token to the hugr)
--UI--
23. Implement a UI for chat interactions
24. Implement a UI for visualizations (tables, charts, maps)

