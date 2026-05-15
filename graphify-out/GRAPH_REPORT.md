# Graph Report - dms-api-server  (2026-05-15)

## Corpus Check
- 118 files · ~10,674 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 692 nodes · 713 edges · 110 communities (77 shown, 33 thin omitted)
- Extraction: 89% EXTRACTED · 11% INFERRED · 0% AMBIGUOUS · INFERRED: 81 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `613a82fd`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 84|Community 84]]
- [[_COMMUNITY_Community 85|Community 85]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]
- [[_COMMUNITY_Community 89|Community 89]]
- [[_COMMUNITY_Community 90|Community 90]]
- [[_COMMUNITY_Community 91|Community 91]]
- [[_COMMUNITY_Community 92|Community 92]]
- [[_COMMUNITY_Community 93|Community 93]]
- [[_COMMUNITY_Community 94|Community 94]]
- [[_COMMUNITY_Community 95|Community 95]]
- [[_COMMUNITY_Community 96|Community 96]]
- [[_COMMUNITY_Community 97|Community 97]]
- [[_COMMUNITY_Community 98|Community 98]]
- [[_COMMUNITY_Community 99|Community 99]]
- [[_COMMUNITY_Community 100|Community 100]]
- [[_COMMUNITY_Community 101|Community 101]]
- [[_COMMUNITY_Community 102|Community 102]]
- [[_COMMUNITY_Community 103|Community 103]]
- [[_COMMUNITY_Community 104|Community 104]]
- [[_COMMUNITY_Community 105|Community 105]]
- [[_COMMUNITY_Community 106|Community 106]]
- [[_COMMUNITY_Community 107|Community 107]]
- [[_COMMUNITY_Community 108|Community 108]]
- [[_COMMUNITY_Community 109|Community 109]]

## God Nodes (most connected - your core abstractions)
1. `NewHandler()` - 11 edges
2. `Knowledge Base` - 11 edges
3. `buildDependencies()` - 9 edges
4. `Service` - 9 edges
5. `NewService()` - 9 edges
6. `SessionRepository` - 9 edges
7. `OK()` - 9 edges
8. `fakeSessionRepo` - 8 edges
9. `OTPRepository` - 8 edges
10. `FromError()` - 8 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewApp()`  [INFERRED]
  cmd/server/main.go → internal/bootstrap/app.go
- `TestHealthRouteShape()` --calls--> `OK()`  [INFERRED]
  tests/smoke/health/health_smoke_test.go → pkg/response/envelope.go
- `TestLogoutRejectsInvalidAccessToken()` --calls--> `NewService()`  [INFERRED]
  tests/unit/auth/service_test.go → internal/modules/auth/service.go
- `TestVerifyOTPRevokesExistingSessionsForSamePlatform()` --calls--> `NewService()`  [INFERRED]
  tests/unit/auth/service_test.go → internal/modules/auth/service.go
- `TestLogoutRequiresAuthorizationHeader()` --calls--> `NewHandler()`  [INFERRED]
  tests/unit/auth/handler_test.go → internal/modules/auth/handler.go

## Communities (110 total, 33 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.09
Nodes (5): fakeOTPRepo, fakeOTPSender, fakeSessionRepo, fakeTokenService, fakeUserRepo

### Community 1 - "Community 1"
Cohesion: 0.15
Nodes (7): newRouter(), contextKey, WithContext(), Recovery(), newRequestID(), RequestID(), RequestLog()

### Community 2 - "Community 2"
Cohesion: 0.08
Nodes (18): getDurationFromSeconds(), getEnv(), getInt(), LoadAuthConfig(), LoadDBConfig(), DBConfig, Connect(), NewPostgresProvider() (+10 more)

### Community 3 - "Community 3"
Cohesion: 0.09
Nodes (11): fakeOTPProvider, fakeOTPRepo, fakeSessionRepo, fakeTokenProvider, fakeUserRepo, NewService(), TestLogoutRejectsInvalidAccessToken(), TestRefreshAndLogout() (+3 more)

### Community 4 - "Community 4"
Cohesion: 0.13
Nodes (14): App, NewApp(), AuthConfig, Config, Load(), MustLoad(), DatabaseConfig, getDurationFromSeconds() (+6 more)

### Community 5 - "Community 5"
Cohesion: 0.14
Nodes (6): NewAuthHandler(), AuthHandler, AuthService, TestLoginBadRequest(), TestVerifyOTPSuccess(), fakeAuthService

### Community 6 - "Community 6"
Cohesion: 0.07
Nodes (16): OTPForType, PlatformType, Repository, toDomain(), UserOTP, OTPFor, OTPPlatform, Repository (+8 more)

### Community 7 - "Community 7"
Cohesion: 0.22
Nodes (3): BaseModel, SoftDeleteableModel, TimestampedModel

### Community 8 - "Community 8"
Cohesion: 0.47
Nodes (7): LoginRequest, LogoutRequest, RefreshTokenRequest, RegisterRequest, TokenResponse, TriggerOTPResponse, VerifyOTPRequest

### Community 10 - "Community 10"
Cohesion: 0.05
Nodes (15): TestAuthLoginRouteShape(), TestAuthRouteContracts(), contractService, fakeAuthService, fakeHandlerAuthService, NewHandler(), TestLoginBadRequest(), TestLogoutRequiresAuthorizationHeader() (+7 more)

### Community 12 - "Community 12"
Cohesion: 0.33
Nodes (4): FuelType, TransmissionType, Vehicle, VehicleType

### Community 14 - "Community 14"
Cohesion: 0.5
Nodes (3): OTPSender, TokenPair, TokenService

### Community 15 - "Community 15"
Cohesion: 0.5
Nodes (3): OTPRepository, SessionRepository, UserRepository

### Community 28 - "Community 28"
Cohesion: 0.07
Nodes (10): OTPRepository, NewOTPRepository(), NewSessionRepository(), SessionRepository, Dependencies, buildDependencies(), NewDummyProvider(), DummyProvider (+2 more)

### Community 29 - "Community 29"
Cohesion: 0.15
Nodes (13): authHeaders, Handler, bindAuthHeaders(), extractBearerToken(), TestHealthRouteShape(), Created(), OK(), Success() (+5 more)

### Community 30 - "Community 30"
Cohesion: 0.13
Nodes (11): Config, OTPFor, OTPPlatform, otpRepo, Service, generateOTPCode(), generateRequestID(), sessionRepo (+3 more)

### Community 31 - "Community 31"
Cohesion: 0.15
Nodes (7): init(), AppError, NewAppError(), Mapper, RegisterMapper(), ToAppError(), init()

### Community 32 - "Community 32"
Cohesion: 0.22
Nodes (3): BaseModel, SoftDeleteableModel, TimestampedModel

### Community 33 - "Community 33"
Cohesion: 0.17
Nodes (11): Architecture Decisions, Conventions, Dependency Rules, How To Use, Important Implementation Details, Important Workflows, Knowledge Base, Known Caveats (+3 more)

### Community 34 - "Community 34"
Cohesion: 0.43
Nodes (4): FuelType, TransmissionType, Vehicle, VehicleType

### Community 35 - "Community 35"
Cohesion: 0.5
Nodes (3): OTPProvider, TokenPair, TokenProvider

### Community 84 - "Community 84"
Cohesion: 0.25
Nodes (7): info, description, name, _postman_id, schema, item, variable

### Community 85 - "Community 85"
Cohesion: 0.25
Nodes (7): info, description, name, _postman_id, schema, item, variable

### Community 86 - "Community 86"
Cohesion: 0.25
Nodes (7): info, description, name, _postman_id, schema, item, variable

### Community 87 - "Community 87"
Cohesion: 0.29
Nodes (6): documentation_policy, notes, required_fields_per_endpoint, required_file_pattern, required_on_change, format

### Community 88 - "Community 88"
Cohesion: 0.33
Nodes (5): Allowed Direction, Cross-Module Dependencies, Dependency Flow, Restricted Direction, Update Rules

### Community 89 - "Community 89"
Cohesion: 0.33
Nodes (5): Architecture Overview, Current Composition Notes, Layers, Purpose, Update Rules

### Community 90 - "Community 90"
Cohesion: 0.33
Nodes (5): Commands, Default Local Configuration, Environment, Local Postgres (Docker), Purpose

### Community 91 - "Community 91"
Cohesion: 0.33
Nodes (5): Auth Module, Boundaries, Documentation Update Checklist, Key Components, Responsibility

### Community 92 - "Community 92"
Cohesion: 0.33
Nodes (5): Boundaries, Customer Module, Documentation Update Checklist, Key Components, Responsibility

### Community 93 - "Community 93"
Cohesion: 0.33
Nodes (5): Boundaries, Documentation Update Checklist, Key Components, Responsibility, Showroom Module

### Community 94 - "Community 94"
Cohesion: 0.33
Nodes (5): Boundaries, Documentation Update Checklist, Key Components, Responsibility, User Module

### Community 95 - "Community 95"
Cohesion: 0.33
Nodes (5): Boundaries, Documentation Update Checklist, Key Components, Responsibility, Vehicle Module

### Community 96 - "Community 96"
Cohesion: 0.33
Nodes (5): Implementations, Interface Ownership, OTP Provider, Responsibility, Update Checklist

### Community 97 - "Community 97"
Cohesion: 0.4
Nodes (4): Architecture Conventions, Change Scope Conventions, Core Principles, Forbidden Patterns

### Community 98 - "Community 98"
Cohesion: 0.4
Nodes (4): Folder Structure, Objective, Placement Rules, Top-Level Placement Guide

### Community 99 - "Community 99"
Cohesion: 0.4
Nodes (4): Governance, Migration Rules, Required Checks, Update Checklist

### Community 100 - "Community 100"
Cohesion: 0.33
Nodes (5): Auth Schema Notes, Module Ownership, Purpose, Schema Overview, Update Checklist

### Community 101 - "Community 101"
Cohesion: 0.4
Nodes (4): Failure Handling, Principles, Transaction Guidelines, Update Checklist

### Community 102 - "Community 102"
Cohesion: 0.4
Nodes (4): Email Provider, Interface Ownership, Responsibility, Update Checklist

### Community 103 - "Community 103"
Cohesion: 0.4
Nodes (4): Interface Ownership, Payment Provider, Responsibility, Update Checklist

### Community 104 - "Community 104"
Cohesion: 0.4
Nodes (4): Interface Ownership, Responsibility, Storage Provider, Update Checklist

### Community 105 - "Community 105"
Cohesion: 0.4
Nodes (4): Clarification Rule, Debugging Workflow, Process, Validation

### Community 106 - "Community 106"
Cohesion: 0.4
Nodes (4): Implementation, Implementation Workflow, Post-Implementation, Pre-Implementation

### Community 107 - "Community 107"
Cohesion: 0.4
Nodes (4): Pre-Release Checks, Release Notes Checklist, Release Workflow, Required Validation Commands

### Community 108 - "Community 108"
Cohesion: 0.4
Nodes (4): Outcome Rules, Required Execution, Scope, Testing Workflow

## Knowledge Gaps
- **155 isolated node(s):** `Dependencies`, `Provider`, `TokenPair`, `Provider`, `SendRequest` (+150 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **33 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `buildDependencies()` connect `Community 28` to `Community 10`, `Community 3`, `Community 4`, `Community 109`?**
  _High betweenness centrality (0.119) - this node is a cross-community bridge._
- **Why does `NewApp()` connect `Community 4` to `Community 1`, `Community 2`, `Community 28`?**
  _High betweenness centrality (0.085) - this node is a cross-community bridge._
- **Why does `NewService()` connect `Community 3` to `Community 28`, `Community 30`?**
  _High betweenness centrality (0.084) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `NewHandler()` (e.g. with `TestAuthLoginRouteShape()` and `TestAuthRouteContracts()`) actually correct?**
  _`NewHandler()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Are the 8 inferred relationships involving `buildDependencies()` (e.g. with `NewApp()` and `NewRepository()`) actually correct?**
  _`buildDependencies()` has 8 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `NewService()` (e.g. with `TestRegisterTriggersOTP()` and `TestVerifyOTPRejectsInvalidCode()`) actually correct?**
  _`NewService()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Dependencies`, `Provider`, `TokenPair` to the rest of the system?**
  _155 weakly-connected nodes found - possible documentation gaps or missing edges._