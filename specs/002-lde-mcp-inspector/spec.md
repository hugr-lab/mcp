# Feature Specification: LDE MCP Inspector

**Feature Branch**: `002-lde-mcp-inspector`
**Created**: 2025-10-30
**Status**: Draft
**Input**: User description: "LDE MCP Inspector
Add the MCP Inspector service to the LDE deployment to testing and validating the MCP functionalities.
MCP Inspector placed in the repository: https://github.com/modelcontextprotocol/inspector
Tasks:
1. Add MCP Inspector service to the docker-compose.yml with port mapping to usage range (in the same range as other services 19000-19020)
2. Add environment variables to the .env.example
3. Add information to the lde readme and in the main readme about the MCP Inspector service in the LDE
4. Add information to the help scripts to setup and start the MCP Inspector service"

## Execution Flow (main)
```
1. Parse user description from Input
   → If empty: ERROR "No feature description provided"
2. Extract key concepts from description
   → Identify: actors, actions, data, constraints
3. For each unclear aspect:
   → Mark with [NEEDS CLARIFICATION: specific question]
4. Fill User Scenarios & Testing section
   → If no clear user flow: ERROR "Cannot determine user scenarios"
5. Generate Functional Requirements
   → Each requirement must be testable
   → Mark ambiguous requirements
6. Identify Key Entities (if data involved)
7. Run Review Checklist
   → If any [NEEDS CLARIFICATION]: WARN "Spec has uncertainties"
   → If implementation details found: ERROR "Remove tech details"
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

### Section Requirements
- **Mandatory sections**: Must be completed for every feature
- **Optional sections**: Include only when relevant to the feature
- When a section doesn't apply, remove it entirely (don't leave as "N/A")

### For AI Generation
When creating this spec from a user prompt:
1. **Mark all ambiguities**: Use [NEEDS CLARIFICATION: specific question] for any assumption you'd need to make
2. **Don't guess**: If the prompt doesn't specify something (e.g., "login system" without auth method), mark it
3. **Think like a tester**: Every vague requirement should fail the "testable and unambiguous" checklist item
4. **Common underspecified areas**:
   - User types and permissions
   - Data retention/deletion policies
   - Performance targets and scale
   - Error handling behaviors
   - Integration requirements
   - Security/compliance needs

---

## Clarifications

### Session 2025-10-30
- Q: What type of data persistence is needed for the Inspector? → A: Full test history - connections + all test operations, results, and session logs
- Q: What authentication approach should the MCP Inspector use? → A: Hybrid - accept both Keycloak tokens OR no auth for localhost access
- Q: Which specific port should be assigned to the MCP Inspector? → A: 19007 - Next sequential port after Hugr health endpoint (19006)
- Q: Should the MCP Inspector test history be included in the standard cleanup or preserved by default? → A: Always remove - test history deleted with all other service data
- Q: What should constitute a "healthy" status for the MCP Inspector? → A: Container running status only - no HTTP check needed

---

## User Scenarios & Testing

### Primary User Story
As a developer working with the Local Development Environment, I need to inspect and debug MCP (Model Context Protocol) server functionality to verify that MCP integrations are working correctly and troubleshoot any issues that arise during development.

### Acceptance Scenarios
1. **Given** the LDE is running, **When** I access the MCP Inspector web interface from localhost, **Then** I should see an interactive UI for testing MCP server capabilities without requiring authentication
2. **Given** the MCP Inspector is accessible remotely, **When** I access it with a valid Keycloak token, **Then** I should be granted access to the Inspector interface
3. **Given** the MCP Inspector is accessible, **When** I configure connection to an MCP server, **Then** I should be able to explore available resources, tools, and prompts
4. **Given** an MCP server is connected, **When** I execute test operations through the Inspector, **Then** I should receive real-time feedback on server responses
5. **Given** I'm developing or debugging MCP servers, **When** I use the Inspector alongside other LDE services, **Then** the Inspector should integrate seamlessly without port conflicts or service interference

### Edge Cases
- What happens when the MCP Inspector service fails to start (port conflicts, resource constraints)?
- How does the system handle connection failures to target MCP servers?
- What if multiple developers try to use the Inspector simultaneously?
- How are authentication credentials managed when connecting to secured MCP servers?
- What happens when a remote user attempts to access without a Keycloak token?

## Requirements

### Functional Requirements
- **FR-001**: System MUST provide a web-based MCP Inspector service accessible via browser
- **FR-002**: System MUST expose the MCP Inspector web interface on port 19007
- **FR-003**: Users MUST be able to start the MCP Inspector as part of the standard LDE startup process
- **FR-004**: Users MUST be able to stop the MCP Inspector when stopping the LDE
- **FR-005**: System MUST verify MCP Inspector health by checking container running status
- **FR-006**: System MUST document the MCP Inspector's purpose, access URL, and usage instructions in LDE documentation
- **FR-007**: Users MUST be able to verify that the MCP Inspector container is running through standard LDE health check scripts
- **FR-008**: System MUST allow configuration of MCP Inspector settings through environment variables
- **FR-009**: System MUST persist MCP Inspector test history across container restarts, including connection configurations, test operations, results, and session logs
- **FR-010**: System MUST support hybrid authentication for MCP Inspector: accept Keycloak OIDC tokens for authenticated access OR allow unauthenticated access from localhost for development convenience
- **FR-011**: System MUST remove all MCP Inspector test history data (connections, operations, results, logs) when cleanup.sh is executed
- **FR-012**: System MUST include MCP Inspector container status check in the standard LDE service health verification process

### Key Entities
- **MCP Inspector Service**: A containerized web application that provides developer tools for testing and debugging MCP servers
- **MCP Inspector Configuration**: Environment settings controlling the Inspector's behavior, port mappings, and connection parameters
- **Test History Data**: Persistent storage containing connection configurations, executed test operations, operation results, and session logs
- **LDE Service Stack**: The collection of services (PostgreSQL, Redis, MinIO, Keycloak, Hugr, MCP Inspector) that comprise the complete development environment
- **Service Documentation**: README files and help scripts that guide users on setup, configuration, and usage

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
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
- [x] Review checklist passed

---
