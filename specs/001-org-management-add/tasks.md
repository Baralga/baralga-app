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
- [x] T001 Create database migration for organizations table in shared/migrations/000007_add_organizations.up.sql
- [x] T002 Create database migration for users table extensions in shared/migrations/000008_add_user_organization.up.sql
- [x] T003 [P] Configure organization domain model structure in user/organization_domain.go

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [ ] T004 [P] Contract test GET /profile/organization in user/organization_web_test.go
- [ ] T005 [P] Contract test POST /profile/organization in user/organization_web_test.go
- [ ] T006 [P] Integration test organization management dialog access in user/organization_web_test.go
- [ ] T007 [P] Integration test organization name update flow in user/organization_web_test.go
- [ ] T008 [P] Unit test organization domain model validation in user/organization_domain_test.go
- [ ] T009 [P] Unit test organization service business logic in user/organization_service_test.go
- [ ] T010 [P] Unit test organization repository interface in user/organization_repository_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [ ] T011 [P] Organization domain model in user/organization_domain.go
- [ ] T012 [P] Organization service interface in user/organization_service.go
- [ ] T013 [P] Organization repository interface in user/organization_repository.go
- [ ] T014 [P] Organization database repository in user/organization_repository_db.go
- [ ] T015 [P] Organization in-memory repository in user/organization_repository_mem.go
- [ ] T016 Organization service implementation in user/organization_service.go
- [ ] T017 Organization web interface GET handler in user/organization_web.go
- [ ] T018 Organization web interface POST handler in user/organization_web.go
- [ ] T019 Organization authorization middleware in user/organization_web.go
- [ ] T020 Organization form validation in user/organization_web.go
- [ ] T021 Organization error handling in user/organization_web.go

## Phase 3.4: Integration
- [ ] T022 Connect organization service to database repository
- [ ] T023 Integrate organization web interface with existing user profile
- [ ] T024 Add organization management link to user profile menu
- [ ] T025 Configure HTMX integration for organization dialog
- [ ] T026 Add CSRF protection to organization form
- [ ] T027 Configure Bootstrap modal for organization dialog
- [ ] T028 Add organization name display to user profile

## Phase 3.5: Polish
- [ ] T029 [P] Unit tests for organization validation in user/organization_domain_test.go
- [ ] T030 [P] Unit tests for organization service in user/organization_service_test.go
- [ ] T031 [P] Integration tests for organization web interface in user/organization_web_test.go
- [ ] T032 [P] Database repository tests in user/organization_repository_db_test.go
- [ ] T033 Performance tests for organization operations
- [ ] T034 Update user documentation for organization management
- [ ] T035 Remove code duplication in organization handlers
- [ ] T036 Run manual testing scenarios from quickstart.md

## Dependencies
- Database migrations (T001-T002) before domain model (T011)
- Domain model (T011) before service (T016)
- Service (T016) before repository implementations (T014-T015)
- Repository implementations (T014-T015) before web interface (T017-T021)
- All core implementation before integration (T022-T028)
- Integration before polish (T029-T036)

## Parallel Example
```
# Launch T004-T010 together (all tests):
Task: "Contract test GET /profile/organization in user/organization_web_test.go"
Task: "Contract test POST /profile/organization in user/organization_web_test.go"
Task: "Integration test organization management dialog access in user/organization_web_test.go"
Task: "Integration test organization name update flow in user/organization_web_test.go"
Task: "Unit test organization domain model validation in user/organization_domain_test.go"
Task: "Unit test organization service business logic in user/organization_service_test.go"
Task: "Unit test organization repository interface in user/organization_repository_test.go"
```

```
# Launch T011-T015 together (domain layer):
Task: "Organization domain model in user/organization_domain.go"
Task: "Organization service interface in user/organization_service.go"
Task: "Organization repository interface in user/organization_repository.go"
Task: "Organization database repository in user/organization_repository_db.go"
Task: "Organization in-memory repository in user/organization_repository_mem.go"
```

```
# Launch T029-T032 together (polish tests):
Task: "Unit tests for organization validation in user/organization_domain_test.go"
Task: "Unit tests for organization service in user/organization_service_test.go"
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

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts**:
   - organization-web.yaml → contract test tasks (T004-T005)
   - GET /profile/organization → implementation task (T017)
   - POST /profile/organization → implementation task (T018)
   
2. **From Data Model**:
   - Organization entity → model creation task (T011)
   - Organization service → service layer task (T016)
   - Organization repository → repository tasks (T014-T015)
   
3. **From User Stories**:
   - Organization management access → integration test (T006)
   - Organization name update → integration test (T007)
   - Validation scenarios → validation tasks (T020)

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
