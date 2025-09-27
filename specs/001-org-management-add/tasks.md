# Tasks: Organization Management Dialog

**Input**: Design documents from `/specs/001-org-management-add/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → If not found: ERROR "No implementation plan found"
   → Extract: tech stack, libraries, structure
2. Load optional design documents:
   → data-model.md: Extract entities → model tasks
   → contracts/: Each file → contract test task
   → research.md: Extract decisions → setup tasks
3. Generate tasks by category:
   → Setup: project init, dependencies, linting
   → Tests: contract tests, integration tests
   → Core: models, services, CLI commands
   → Integration: DB, middleware, logging
   → Polish: unit tests, performance, docs
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All contracts have tests?
   → All entities have models?
   → All endpoints implemented?
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Go project**: `user/`, `shared/`, `auth/`, `tracking/` modules
- **Database**: `shared/migrations/` for schema changes
- **Tests**: Co-located with implementation files

## Phase 3.1: Setup
- [x] T001 Use existing Principal roles instead of creating new database tables
- [x] T002 Extend existing OrganizationRepository interface with new methods in user/organization_repository.go
- [x] T003 [P] Create organization service interface in user/organization_service.go

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] T004 [P] Contract test GET /profile/organization in user/organization_web_test.go
- [x] T005 [P] Contract test POST /profile/organization in user/organization_web_test.go
- [x] T006 [P] Integration test organization management dialog access in user/organization_web_test.go
- [x] T007 [P] Integration test organization title update flow in user/organization_web_test.go
- [x] T008 [P] Unit test organization service business logic in user/organization_service_test.go
- [x] T009 [P] Unit test organization repository interface in user/organization_repository_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [x] T010 [P] Organization service implementation in user/organization_service.go
- [x] T011 [P] Extend database repository with new methods in user/organization_repository_db.go
- [x] T012 [P] Extend in-memory repository with new methods in user/organization_repository_mem.go
- [x] T013 Organization web interface GET handler in user/organization_web.go
- [x] T014 Organization web interface POST handler in user/organization_web.go
- [x] T015 Organization authorization middleware in user/organization_web.go
- [x] T016 Organization form validation in user/organization_web.go
- [x] T017 Organization error handling in user/organization_web.go
- [x] T025 Convert organization web handlers to use gomponents instead of raw HTML
- [x] T026 Integrate organization modal with global modal system like activity web handlers
- [x] T027 Update organization service to use repositoryTxer.InTx for database transactions

## Phase 3.4: Integration
- [x] T018 Connect organization service to database repository
- [x] T019 Integrate organization web interface with navbar dropdown
- [x] T020 Add organization management link to navbar dropdown menu
- [x] T021 Configure HTMX integration for organization dialog
- [x] T022 Add CSRF protection to organization form
- [x] T023 Configure Bootstrap modal for organization dialog
- [x] T024 Add organization title display to user profile

## Phase 3.5: Polish
- [x] T025 [P] Unit tests for organization validation in user/organization_service_test.go
- [x] T026 [P] Integration tests for organization web interface in user/organization_web_test.go
- [x] T027 [P] Database repository tests in user/organization_repository_db_test.go
- [x] T028 Performance tests for organization operations
- [x] T029 Update user documentation for organization management
- [x] T030 Remove code duplication in organization handlers
- [x] T031 Run manual testing scenarios from quickstart.md

## Dependencies
- Database migration (T001) before repository extensions (T011-T012)
- Repository interface extension (T002) before repository implementations (T011-T012)
- Service interface (T003) before service implementation (T010)
- Service implementation (T010) before web interface (T013-T017)
- All core implementation before integration (T018-T024)
- Integration before polish (T025-T031)

## Parallel Example
```
# Launch T004-T009 together (all tests):
Task: "Contract test GET /profile/organization in user/organization_web_test.go"
Task: "Contract test POST /profile/organization in user/organization_web_test.go"
Task: "Integration test organization management dialog access in user/organization_web_test.go"
Task: "Integration test organization title update flow in user/organization_web_test.go"
Task: "Unit test organization service business logic in user/organization_service_test.go"
Task: "Unit test organization repository interface in user/organization_repository_test.go"
```

```
# Launch T010-T012 together (service and repository implementations):
Task: "Organization service implementation in user/organization_service.go"
Task: "Extend database repository with new methods in user/organization_repository_db.go"
Task: "Extend in-memory repository with new methods in user/organization_repository_mem.go"
```

```
# Launch T025-T027 together (polish tests):
Task: "Unit tests for organization validation in user/organization_service_test.go"
Task: "Integration tests for organization web interface in user/organization_web_test.go"
Task: "Database repository tests in user/organization_repository_db_test.go"
```

## Notes
- [P] tasks = different files, no dependencies
- Verify tests fail before implementing
- Commit after each task
- Avoid: vague tasks, same file conflicts
- Follow DDD layered architecture
- Use existing HTMX/Bootstrap patterns
- Maintain consistency with existing user module
- Leverage existing organization infrastructure

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts**:
   - organization-web.yaml → contract test tasks (T004-T005)
   - GET /profile/organization → implementation task (T013)
   - POST /profile/organization → implementation task (T014)
   
2. **From Data Model**:
   - Organization entity → service layer task (T010)
   - Repository interface extension → repository tasks (T011-T012)
   
3. **From User Stories**:
   - Organization management access → integration test (T006)
   - Organization title update → integration test (T007)
   - Validation scenarios → validation tasks (T016)

4. **Ordering**:
   - Setup → Tests → Models → Services → Endpoints → Polish
   - Dependencies block parallel execution

## Validation Checklist
*GATE: Checked by main() before returning*

- [x] All contracts have corresponding tests
- [x] All entities have model tasks
- [x] All tests come before implementation
- [x] Parallel tasks truly independent
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task