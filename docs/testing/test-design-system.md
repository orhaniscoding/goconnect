# System-Level Test Design

## Testability Assessment

- **Controllability**: **CONCERNS**
  - **Issue**: WireGuard interface management requires `root/sudo` privileges.
  - **Impact**: CI environments and local dev testing require elevated permissions or robust mocking.
  - **Mitigation**: Implement a `WireGuardInterface` interface to allow dependency injection of a mock implementation for unit/integration tests. Use privileged runners only for E2E.
  - **Positive**: API-driven architecture (REST/gRPC) allows for easy triggering of flows (e.g., Device Flow) without complex UI interaction.

- **Observability**: **PASS**
  - **Details**: Go's standard logging and metrics libraries can provide visibility.
  - **Requirement**: ensuring structured JSON logging is implemented across Core and CLI to trace requests (Trace IDs) is critical for debugging distributed auth flows.

- **Reliability**: **PASS**
  - **Details**: Stateless API design (utilizing Redis for session state) supports parallel test execution.
  - **Risk**: Polling mechanisms (Device Flow) in tests can lead to flakiness if implemented with hardcoded sleeps. Use deterministic wait helpers or channel-based synchronization in tests.

## Architecturally Significant Requirements (ASRs)

| ASR ID | Requirement | Probability | Impact | Score | Category |
|--------|-------------|-------------|--------|-------|----------|
| ASR-01 | **Secure Headless Authentication**: Device Flow must work reliably without GUI interventions. | 3 (Likely failure in complex net) | 3 (Critical - no access) | **9** | SEC/BUS |
| ASR-02 | **WireGuard Tunnel Stability**: Tunnel must persist/recover across network changes. | 2 (Possible) | 3 (Critical) | **6** | REL/TECH |
| ASR-03 | **Cross-Platform Keyring**: Secure token storage must work on Linux, macOS, Windows. | 2 (Possible) | 2 (Degraded) | **4** | SEC/TECH |
| ASR-04 | **Low Latency Overhead**: Daemon must add <5ms overhead to connection setup. | 2 (Possible) | 2 (Degraded) | **4** | PERF |

## Test Levels Strategy

Given the system architecture (Go Backend + CLI + Desktop), we recommend a **Pyramid** approach:

- **Unit Tests: 60%**
  - **Scope**: Core logic, Configuration generation, State machines, Utility functions.
  - **Tools**: Go `testing`, `testify`, `mockery`.
  - **Rationale**: Fastest feedback, handles complex permutations of config/state without sudo.

- **Integration Tests: 30%**
  - **Scope**: API Handlers + DB, Service Layer + Redis, CLI Command logic + Mock Server.
  - **Tools**: Go `testing` with Docker containers (testcontainers-go).
  - **Rationale**: Validates contracts and state persistence. Essential for Auth flows.

- **E2E Tests: 10%**
  - **Scope**: Full "Login -> Connect -> Disconnect" flows.
  - **Tools**: `testscript` (CLI), specific E2E suites running on privileged VMs.
  - **Rationale**: Slow, brittle, requires expensive setup. Use sparingly for critical paths (ASR-01, ASR-02).

## NFR Testing Approach

- **Security (SEC)**:
  - **Automated**: Static Analysis (`gosec`), Dependency Scanning (`govulncheck`).
  - **Functional**: E2E tests for Auth flows (Happy path & Error cases like Expired Token).
  - **Secret Handling**: Unit tests verifying `Keyring` abstraction sanitizes logs and encrypts data.

- **Performance (PERF)**:
  - **Load Testing**: `k6` scripts targeting the Auth API (`/v1/auth/device/code`) and Signaling endpoints.
  - **Benchmarking**: Go benchmarks (`go test -bench`) for hot paths in packet processing/config generation.

- **Reliability (REL)**:
  - **Chaos**: Integration tests that simulate Redis unavailability to verify graceful degradation.
  - **Recovery**: "Burn-in" tests for the implementation of the Polling mechanism to ensure it doesn't leak routines or hang.

- **Maintainability (TECH)**:
  - **Coverage**: Enforce >80% code coverage on `core/` modules.
  - **Linting**: `golangci-lint` with strict settings (including `errcheck`, `gocyclo`).

## Test Environment Requirements

- **Local Dev**:
  - Docker Compose (Postgres, Redis).
  - Check for `sudo` capability if running Integration tests involving real interfaces.

- **CI/CD**:
  - **Standard Runner**: Unit tests, linting, build.
  - **Privileged Runner**: Integration/E2E tests needing `CAP_NET_ADMIN` (WireGuard).

## Testability Concerns

- **Privileged E2E**: Running `wg-quick` or equivalent requires root. This complicates CI pipelines and local testing on restricted machines.
  - *Recommendation*: Use mocks for 90% of testing; restrict real interface creation to a dedicated "Nightly" or "Privileged" CI job.

## Recommendations for Sprint 0

1. **Scaffold Test Architecture**:
   - Initialize `testcontainers-go` setup for Integration tests.
   - Set up `mockery` for interface generation.
2. **Implement WireGuard Mock**: Create a `Netlink` interface wrapper to mock kernel interactions.
3. **CI Config**: Configure GitHub Actions to separate "Unit" (fast) from "Integration" (slow/privileged) jobs.
