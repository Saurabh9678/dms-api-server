# Graph Report - dms-api-server  (2026-05-15)

## Corpus Check
- 91 files · ~7,079 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 517 nodes · 552 edges · 84 communities (51 shown, 33 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 75 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `32734877`
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

## God Nodes (most connected - your core abstractions)
1. `buildDependencies()` - 9 edges
2. `Service` - 9 edges
3. `OK()` - 9 edges
4. `NewHandler()` - 8 edges
5. `FromError()` - 8 edges
6. `contractService` - 7 edges
7. `newRouter()` - 7 edges
8. `Handler` - 7 edges
9. `NewService()` - 7 edges
10. `OTPRepository` - 7 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `BuildServer()`  [INFERRED]
  cmd/server/main.go → internal/wire/server.go
- `TestHealthRouteShape()` --calls--> `OK()`  [INFERRED]
  tests/smoke/health/health_smoke_test.go → pkg/response/envelope.go
- `newRouter()` --calls--> `Recovery()`  [INFERRED]
  internal/bootstrap/router.go → pkg/middleware/recovery.go
- `main()` --calls--> `NewApp()`  [INFERRED]
  cmd/server/main.go → internal/bootstrap/app.go
- `TestAuthLoginRouteShape()` --calls--> `RegisterRoutes()`  [INFERRED]
  tests/smoke/auth/auth_routes_smoke_test.go → internal/modules/auth/routes.go

## Communities (84 total, 33 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.09
Nodes (5): fakeOTPRepo, fakeOTPSender, fakeSessionRepo, fakeTokenService, fakeUserRepo

### Community 1 - "Community 1"
Cohesion: 0.13
Nodes (11): Config, OTPFor, OTPPlatform, otpRepo, Service, generateOTPCode(), sessionRepo, UserOTP (+3 more)

### Community 2 - "Community 2"
Cohesion: 0.09
Nodes (17): getDurationFromSeconds(), getEnv(), getInt(), LoadAuthConfig(), LoadDBConfig(), DBConfig, Connect(), NewPostgresProvider() (+9 more)

### Community 3 - "Community 3"
Cohesion: 0.11
Nodes (9): fakeOTPProvider, fakeOTPRepo, fakeSessionRepo, fakeTokenProvider, fakeUserRepo, NewService(), TestRefreshAndLogout(), TestRegisterTriggersOTP() (+1 more)

### Community 4 - "Community 4"
Cohesion: 0.11
Nodes (15): App, NewApp(), AuthConfig, Config, Load(), MustLoad(), DatabaseConfig, getDurationFromSeconds() (+7 more)

### Community 5 - "Community 5"
Cohesion: 0.14
Nodes (6): NewAuthHandler(), AuthHandler, AuthService, TestLoginBadRequest(), TestVerifyOTPSuccess(), fakeAuthService

### Community 6 - "Community 6"
Cohesion: 0.2
Nodes (4): Repository, toDomain(), SessionPlatformType, UserSession

### Community 7 - "Community 7"
Cohesion: 0.22
Nodes (3): BaseModel, SoftDeleteableModel, TimestampedModel

### Community 8 - "Community 8"
Cohesion: 0.47
Nodes (7): LoginRequest, LogoutRequest, RefreshTokenRequest, RegisterRequest, TokenResponse, TriggerOTPResponse, VerifyOTPRequest

### Community 10 - "Community 10"
Cohesion: 0.05
Nodes (12): TestAuthLoginRouteShape(), TestAuthRouteContracts(), contractService, fakeAuthService, fakeHandlerAuthService, NewHandler(), TestLoginBadRequest(), TestVerifyOTPSuccess() (+4 more)

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
Cohesion: 0.08
Nodes (10): OTPRepository, NewOTPRepository(), NewSessionRepository(), SessionRepository, Dependencies, buildDependencies(), NewDummyProvider(), DummyProvider (+2 more)

### Community 29 - "Community 29"
Cohesion: 0.09
Nodes (17): Handler, newRouter(), TestHealthRouteShape(), contextKey, WithContext(), Recovery(), newRequestID(), RequestID() (+9 more)

### Community 30 - "Community 30"
Cohesion: 0.12
Nodes (11): OTPForType, PlatformType, Repository, toDomain(), UserOTP, OTPFor, UserEntity, UserOTPEntity (+3 more)

### Community 31 - "Community 31"
Cohesion: 0.15
Nodes (7): init(), AppError, NewAppError(), Mapper, RegisterMapper(), ToAppError(), init()

### Community 32 - "Community 32"
Cohesion: 0.22
Nodes (3): BaseModel, SoftDeleteableModel, TimestampedModel

### Community 34 - "Community 34"
Cohesion: 0.43
Nodes (4): FuelType, TransmissionType, Vehicle, VehicleType

### Community 35 - "Community 35"
Cohesion: 0.5
Nodes (3): OTPProvider, TokenPair, TokenProvider

## Knowledge Gaps
- **48 isolated node(s):** `Dependencies`, `Provider`, `TokenPair`, `Provider`, `SendRequest` (+43 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **33 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `buildDependencies()` connect `Community 28` to `Community 33`, `Community 10`, `Community 3`, `Community 4`?**
  _High betweenness centrality (0.187) - this node is a cross-community bridge._
- **Why does `NewApp()` connect `Community 4` to `Community 28`, `Community 29`?**
  _High betweenness centrality (0.143) - this node is a cross-community bridge._
- **Why does `NewService()` connect `Community 3` to `Community 1`, `Community 28`?**
  _High betweenness centrality (0.131) - this node is a cross-community bridge._
- **Are the 8 inferred relationships involving `buildDependencies()` (e.g. with `NewApp()` and `NewRepository()`) actually correct?**
  _`buildDependencies()` has 8 INFERRED edges - model-reasoned connections that need verification._
- **Are the 7 inferred relationships involving `OK()` (e.g. with `TestHealthRouteShape()` and `newRouter()`) actually correct?**
  _`OK()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `NewHandler()` (e.g. with `TestAuthLoginRouteShape()` and `TestAuthRouteContracts()`) actually correct?**
  _`NewHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `FromError()` (e.g. with `.Register()` and `.Login()`) actually correct?**
  _`FromError()` has 6 INFERRED edges - model-reasoned connections that need verification._