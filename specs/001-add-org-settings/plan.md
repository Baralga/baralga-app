# Implementation Plan: Organization Management Dialog

**Branch**: `001-add-org-settings` | **Date**: 2024-12-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-add-org-settings/spec.md`

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
Add organization management dialog accessible from user profile for admin users. Admins can modify organization name through a modal dialog with proper validation and authorization checks. The feature integrates with existing DDD architecture using Go, PostgreSQL, HTMX, and Bootstrap components.

## Technical Context
**Language/Version**: Go 1.24+ with toolchain go1.24.5  
**Primary Dependencies**: Chi router (go-chi/chi/v5), HTMX 2.0.6, Bootstrap 5.3.2, gomponents for HTML generation  
**Storage**: PostgreSQL with pgx/v5 driver  
**Testing**: testify/assert, matryer/is, ory/dockertest/v3 for integration tests  
**Target Platform**: Web application with server-side rendered HTML  
**Project Type**: Web application (backend + frontend)  
**Performance Goals**: Standard web response times (<200ms for form submissions)  
**Constraints**: Must follow DDD layered architecture, maintain existing security patterns  
**Scale/Scope**: Single organization per user, admin-only access

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Architecture Compliance
- [x] **Domain-Driven Design**: Feature follows existing module structure (user/ module)
- [x] **Layered Dependencies**: Presentation → Domain → Infrastructure (no violations)
- [x] **Module Boundaries**: Organization management belongs in user/ module
- [x] **Clean Architecture**: Clear separation between web handlers, services, and repositories
- [x] **Test-First**: Domain logic will be unit tested, infrastructure uses existing patterns

### Code Organization
- [x] **Module Structure**: Extends existing user/ module with organization management
- [x] **Repository Pattern**: Uses existing OrganizationRepository interface
- [x] **Service Layer**: Business logic in user service layer
- [x] **Web Handlers**: Presentation layer in user_web.go

## Project Structure

### Documentation (this feature)
```
specs/001-add-org-settings/
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
├── user_domain.go           # Add organization update methods to Organization struct
├── user_service.go          # Add UpdateOrganization method
├── user_web.go             # Add organization management handlers and forms
├── organization_repository_db.go    # Add UpdateOrganization method
└── organization_repository_mem.go  # Add UpdateOrganization method

shared/
├── migrations/             # Add organization name update migration if needed
└── assets/                # No changes needed (uses existing modal.js)

auth/
└── (no changes - uses existing role-based authorization)
```

**Structure Decision**: Extends existing user module with organization management capabilities. No new modules required. Uses existing modal system and form patterns.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - Organization name validation rules and constraints
   - Modal dialog integration with existing HTMX patterns
   - Authorization flow for admin-only access
   - Database migration requirements for organization updates

2. **Generate and dispatch research agents**:
   ```
   Task: "Research organization name validation patterns in Go web applications"
   Task: "Research HTMX modal integration patterns with Bootstrap 5"
   Task: "Research role-based authorization patterns in existing codebase"
   Task: "Research database update patterns for organization entities"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Organization entity with name field and validation rules
   - User entity with admin role checking
   - Form models for organization management

2. **Generate API contracts** from functional requirements:
   - GET /profile/organization - Show organization management dialog
   - POST /profile/organization - Update organization name
   - Output OpenAPI schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - Organization management endpoint tests
   - Authorization tests for admin-only access
   - Validation tests for organization name updates

4. **Extract test scenarios** from user stories:
   - Admin can access organization management
   - Non-admin cannot access organization management
   - Organization name validation and update flow

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh cursor`
   - Add organization management context
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency

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

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (No violations detected) | | |

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [ ] Phase 0: Research complete (/plan command)
- [ ] Phase 1: Design complete (/plan command)
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [ ] Post-Design Constitution Check: PASS
- [ ] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
