# Feature Specification: Local Development Environment for Hugr MCP Service

**Feature Branch**: `001-create-a-local`
**Created**: 2025-10-27
**Status**: Draft
**Input**: User description: "Create a local development environment for Hugr MCP service with necessary dependencies and example data. Use the Docker compose for examples environment to create the new docker deployment setup dependencies service like Keycloak, hugr, Postgres, Redis, MinIO. The example data can be generated using hugr github repository - https://github.com/hugr/examples, the local clone of the repo is /Users/vgribanov/projects/hugr-lab/examples. Use the Open payments data for 2023 and Synthea (generate Synthea patient data) to fill the hugr with relevant healthcare data (use the /Users/vgribanov/projects/synthea project to generate data.). Create shell scripts to start and stop the environment, and to fill the hugr with example data."

## Execution Flow (main)
```
1. Parse user description from Input
   → Feature description provided ✓
2. Extract key concepts from description
   → Actors: Developers working on Hugr MCP service
   → Actions: Start environment, stop environment, load example data
   → Data: Open Payments 2023, Synthea healthcare patient data
   → Constraints: Must use specific data sources and local repositories
3. For each unclear aspect:
4. Fill User Scenarios & Testing section ✓
5. Generate Functional Requirements ✓
6. Identify Key Entities ✓
7. Run Review Checklist
   → Major uncertainties resolved
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-27
- Q: What scale of Synthea patient data should be generated for the local development environment? → A: 10000 patients
- Q: Should the environment include a reset/clean command to start with a fresh state? → A: Yes, manual reset flag; volumes in .local directory
- Q: What Keycloak authentication configuration is required for the local development environment? → A: Realm with admin, viewer, analyst roles from hugr examples
- Q: What data sources should be created in Hugr for the loaded datasets? → A: Synthea (DuckDB) and Open Payments
- Q: What service versions should be used in the local development environment? → A: Latest as in hugr examples

---

## User Scenarios & Testing

### Primary User Story
As a developer working on the Hugr MCP service, I need a consistent and reproducible local development environment that includes all required services (Hugr, Keycloak, Postgres, Redis, MinIO) pre-configured and loaded with realistic healthcare data so that I can develop and test MCP features without setting up each dependency manually or relying on remote environments.

### Acceptance Scenarios
1. **Given** no local environment exists, **When** developer runs the start script, **Then** all required services are running and accessible, and the developer can interact with Hugr containing healthcare data
2. **Given** the environment is running, **When** developer runs the stop script, **Then** all services are gracefully stopped and can be restarted without data loss
3. **Given** the environment is running with empty Hugr instance, **When** developer runs the data loading script, **Then** Hugr contains Open Payments 2023 data and generated Synthea patient data accessible for queries
4. **Given** a clean system, **When** developer runs the complete setup for the first time, **Then** environment starts successfully with all services authenticated and data sources properly configured

### Edge Cases
- What happens when required external data sources (examples repository, Synthea generator) are missing or inaccessible?
- How does the system handle partial failures (e.g., one service fails to start)?
- What happens when data already exists in Hugr and the data loading script is run again?
- How does the developer verify that all services are healthy and properly connected?
- What happens when ports are already in use by other services?
- How can developers reset the environment to a clean state? (Answered: Manual reset flag wipes volumes in .local directory)

## Requirements

### Functional Requirements
- **FR-001**: System MUST provide a single command to start all required services (Hugr, Keycloak, Postgres, Redis, MinIO) in a local development environment
- **FR-002**: System MUST provide a single command to stop all running services gracefully
- **FR-003**: System MUST load Open Payments data for year 2023 from the local examples repository into Hugr
- **FR-004**: System MUST generate Synthea patient healthcare data in DuckDB format using the local Synthea project and load it into Hugr
- **FR-005**: System MUST configure Keycloak realm with at least three user roles (admin, viewer, analyst) using the configuration pattern from the hugr examples project
- **FR-006**: System MUST create data sources in Hugr via GraphQL mutations for loaded datasets
- **FR-007**: System MUST persist data across environment restarts (Postgres data, MinIO objects, Keycloak configuration) in a .local directory
- **FR-007a**: System MUST store all service volumes (pg-data, minio, etc.) in .local directory following the same pattern as the hugr examples project
- **FR-008**: System MUST be reproducible - running setup on a different machine should yield identical environment
- **FR-009**: System MUST validate that required dependencies (Docker, example repositories, Synthea) are available before starting
- **FR-010**: System MUST provide clear feedback on the status of each service during startup and data loading
- **FR-011**: System MUST handle graceful shutdown allowing developers to restart without corrupting state
- **FR-011a**: System MUST support a manual reset flag that wipes all data volumes and recreates a clean environment when explicitly provided
- **FR-011b**: System MUST preserve existing data volumes when reset flag is not provided, allowing developers to continue with last known state
- **FR-012**: System MUST use service versions matching those specified in the hugr examples project configuration
- **FR-013**: System MUST configure service networking so all services can communicate with each other
- **FR-014**: Developers MUST be able to access Hugr GraphQL endpoint from their host machine [NEEDS CLARIFICATION: What ports should be exposed?]
- **FR-015**: System MUST generate Synthea patient data with 10000 patients
- **FR-016**: System MUST create two data sources in Hugr: Synthea (DuckDB format) and Open Payments (as configured in hugr examples project)

### Key Entities
- **Local Development Environment**: Container-based environment encompassing all services needed for Hugr MCP development, isolated from production and other environments
- **Service Dependencies**: Set of required services (Hugr instance, Keycloak identity provider, Postgres database, Redis cache, MinIO object storage) that must run and communicate with each other
- **Example Healthcare Datasets**: Two primary datasets - Open Payments 2023 data (financial healthcare payments) and Synthea generated patient data (synthetic healthcare records)
- **Data Sources**: Hugr entities representing connections to different datasets, created via GraphQL mutations and configured with appropriate metadata
- **Environment Control Scripts**: Executables that manage environment lifecycle (start, stop, data loading) providing developers with simple interface to complex multi-service orchestration

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain (5 answered, 1 deferred)
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed (major clarifications complete)

---
