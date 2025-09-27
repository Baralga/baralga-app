# Implementation Plan: Organization Management Dialog

**Branch**: `001-org-management-add` | **Date**: 2024-12-19 | **Spec**: [link]
**Input**: Feature specification from `/specs/001-org-management-add/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
**Primary Requirement**: Add organization management dialog accessible from user profile for administrators to change organization name.

**Technical Approach**: Extend existing user module with organization management capabilities. The organization infrastructure already exists (tables, repositories, domain models) but needs web interface and service methods for name updates. Focus on web layer implementation using existing HTMX/Bootstrap patterns.

## Technical Context
**Language/Version**: Go 1.24+ with toolchain go1.24.5  
**Primary Dependencies**: Chi router (go-chi/chi/v5), HTMX 2.0.6, Bootstrap 5.3.2, gomponents for templating  
**Storage**: PostgreSQL with pgx/v5 driver  
**Testing**: testify/assert, matryer/is, ory/dockertest/v3 for integration tests  
**Target Platform**: Web application with server-side rendered HTML  
**Project Type**: Web application (Go backend with HTMX frontend)  
**Performance Goals**: Standard web application performance targets  
**Constraints**: Must follow DDD layered architecture, no API endpoints (web-only), use existing organization infrastructure  
**Scale/Scope**: Organization management for existing user base

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### DDD Architecture Compliance
- [x] **Layered Architecture**: Feature will follow presentation → domain → infrastructure layers
- [x] **Module Organization**: Will extend existing user module for organization management
- [x] **Dependency Rules**: Presentation layer will depend only on domain layer
- [x] **Domain Isolation**: Business logic will be in domain services, not in presentation layer

### Module Structure Compliance  
- [x] **Existing Module**: Will extend user module with organization management capabilities
- [x] **Cross-Module Dependencies**: No new cross-module dependencies introduced
- [x] **Shared Module Independence**: No changes to shared module dependencies

### Clean Architecture Boundaries
- [x] **Presentation Layer**: Web interface using existing HTMX/Bootstrap patterns
- [x] **Domain Layer**: Extend existing organization service for business logic
- [x] **Infrastructure Layer**: Use existing organization repository implementations

### Test-First Development
- [x] **Domain Testing**: Organization service will have comprehensive unit tests
- [x] **Integration Testing**: Web interface will have integration tests
- [x] **Repository Testing**: Database repository will be tested with integration tests

## Project Structure

### Documentation (this feature)
```
specs/001-org-management-add/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
user/
├── organization_domain.go          # EXISTING - Organization domain model
├── organization_service.go          # NEW - Organization service for business logic
├── organization_service_test.go     # NEW - Unit tests for organization service
├── organization_repository_db.go    # EXISTING - Database repository
├── organization_repository_mem.go   # EXISTING - In-memory repository
├── organization_web.go             # NEW - Web interface for organization management
├── organization_web_test.go        # NEW - Integration tests for web interface
└── user_web.go                     # EXISTING - User web handlers (to be extended)

shared/
├── assets/                         # EXISTING - Static assets (Bootstrap, HTMX, etc.)
├── migrations/                     # EXISTING - Database migrations
└── [existing shared utilities]

auth/
└── [existing authentication module]

tracking/
└── [existing tracking module]
```

**Structure Decision**: Extend existing user module with organization management capabilities. The organization infrastructure already exists (domain model, repositories) but needs service layer and web interface. Focus on minimal changes to existing code while adding new organization management functionality.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh cursor`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P] 
- Each user story → integration test task
- Implementation tasks to make tests pass

**Ordering Strategy**:
- TDD order: Tests before implementation 
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 15-20 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

**Specific Task Categories for Organization Management**:
1. **Service Layer**: Organization service for business logic and name updates
2. **Presentation Layer**: Web interface, HTMX integration, Bootstrap modal
3. **Testing**: Unit tests for organization service, integration tests for web interface
4. **Security**: Authorization checks, input validation, CSRF protection

**Task Dependencies**:
- Organization service must exist before web interface
- Web interface must exist before integration tests
- All components must exist before polish tasks

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Web interface addition | Required for user interaction | Direct API calls insufficient for user experience |

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*