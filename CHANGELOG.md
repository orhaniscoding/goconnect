# Changelog

## [2.0.0](https://github.com/orhaniscoding/goconnect/compare/v1.2.0...v2.0.0) (2025-11-20)


### ⚠ BREAKING CHANGES

* **server:** Default storage backend changed from in-memory to PostgreSQL
* **tests:** AuthService.Register now returns AuthResponse instead of User

### Features

* Add IP Allocation Display UI ([f00a049](https://github.com/orhaniscoding/goconnect/commit/f00a0493cde4515843706aaab133aaa1e42abe61))
* Add QR code generation for WireGuard configs ([a4e8e22](https://github.com/orhaniscoding/goconnect/commit/a4e8e22c3cab0eb91c8ceb039ed4142d3504cf3a))
* Add WireGuard config download UI ([d5cd0fd](https://github.com/orhaniscoding/goconnect/commit/d5cd0fda6ac968a1484f84222970218f45c2d98b))
* **auth:** implement production-ready JWT authentication ([dda2132](https://github.com/orhaniscoding/goconnect/commit/dda2132f1d28823e2697702437ff2c2db4646783))
* **dev:** add comprehensive Makefile system and developer documentation ([d626153](https://github.com/orhaniscoding/goconnect/commit/d6261534b4e6512176150fb7daf6b2d83a5d56f6))
* implement join requests list endpoint and UI ([95c35a3](https://github.com/orhaniscoding/goconnect/commit/95c35a31957a0d99258654fc431ad07470a00e9c))
* **server:** implement PostgreSQL persistence layer ([cd188bd](https://github.com/orhaniscoding/goconnect/commit/cd188bd5334da30212305384d291038911f76ea3))
* **web-ui:** implement authentication guard for protected routes ([761c0d3](https://github.com/orhaniscoding/goconnect/commit/761c0d33e9967596a45dd66530f9f72c61b259c0))
* **web-ui:** implement device management UI ([2d4c11d](https://github.com/orhaniscoding/goconnect/commit/2d4c11d120af2ca3e6e1b70422c1323c5335a279))
* **web-ui:** implement login page with JWT authentication ([047c4d0](https://github.com/orhaniscoding/goconnect/commit/047c4d0b7169a5508d8671d583226dad65cb99db))
* **web-ui:** implement network details page with membership management ([b89f415](https://github.com/orhaniscoding/goconnect/commit/b89f415ea76e69ced1f4f676f718b28e6fd0f74f))
* **web-ui:** implement network management UI with CRUD operations ([43cb5ab](https://github.com/orhaniscoding/goconnect/commit/43cb5ab041fb6e2acc55991badcdeaa4c26457af))
* **web-ui:** implement registration page with validation ([20b1eb6](https://github.com/orhaniscoding/goconnect/commit/20b1eb65b9fa6087887db843f79cfbaa8c97362f))


### Bug Fixes

* **tests:** resolve all handler test failures and parameter order bugs ([e53ed9d](https://github.com/orhaniscoding/goconnect/commit/e53ed9d2f0926e96f016430bce557de3a8375a30))

## [1.2.0](https://github.com/orhaniscoding/goconnect/compare/v1.1.0...v1.2.0) (2025-11-20)


### Features

* implement comprehensive authentication, chat, device and peer management system ([0a64ea7](https://github.com/orhaniscoding/goconnect/commit/0a64ea7b109ec97222a4b878cc4c73ad90c7aea7))
* **server:** implement authentication system phase 1 ([#71](https://github.com/orhaniscoding/goconnect/issues/71)) ([ef9e5d8](https://github.com/orhaniscoding/goconnect/commit/ef9e5d8c9f097a4502f266b44dab7f4533f5d902))


### Bug Fixes

* add tenant_id parameter to network handler service calls ([3516b74](https://github.com/orhaniscoding/goconnect/commit/3516b74bd688992c2add13b66ef215d730411c04))
* enforce tenant isolation in network repository list operations ([e9ccca9](https://github.com/orhaniscoding/goconnect/commit/e9ccca9eb0db02681d88a0b99b3936a994e88907))
* resolve merge conflicts and unify codebase ([9c45145](https://github.com/orhaniscoding/goconnect/commit/9c45145a8d6d4b452d24399eb82b4589da8df568))
* update test files for PassHash to PasswordHash migration and tenant_id parameters ([3b7e420](https://github.com/orhaniscoding/goconnect/commit/3b7e42074ec2e13a3084f0e86e9fdbacfac868ed))

## [1.1.0](https://github.com/orhaniscoding/goconnect/compare/v1.0.0...v1.1.0) (2025-10-10)


### Features

* **audit:** persistence, integrity chain, rotation, signing ([#50](https://github.com/orhaniscoding/goconnect/issues/50)) ([080d539](https://github.com/orhaniscoding/goconnect/commit/080d53997bbe94725696a5a876786309a035560e))
* **ipam:** add allocations listing endpoint and tests ([#35](https://github.com/orhaniscoding/goconnect/issues/35)) ([53acc79](https://github.com/orhaniscoding/goconnect/commit/53acc793dd5703fe0ac240345fa44255a13d973c))
* **ipam:** add base IPAM domain, repository, service, tests and spec placeholder ([#30](https://github.com/orhaniscoding/goconnect/issues/30)) ([b1c6fce](https://github.com/orhaniscoding/goconnect/commit/b1c6fce18361826a7e5c9a6c814fa33192d339af))
* **ipam:** add IP release endpoint, service logic, offset reuse and tests ([#37](https://github.com/orhaniscoding/goconnect/issues/37)) ([730b3db](https://github.com/orhaniscoding/goconnect/commit/730b3dbcfdcc0609f81d15c844a7841b6c1291bb))
* **ipam:** admin/owner release endpoint, RBAC service method, tests, concurrency stress test and spec update ([#39](https://github.com/orhaniscoding/goconnect/issues/39)) ([8992126](https://github.com/orhaniscoding/goconnect/commit/8992126db3275b1645d2c2f0eb74615ba34196e4))
* **ipam:** enforce membership for IP allocation and add tests ([#34](https://github.com/orhaniscoding/goconnect/issues/34)) ([b4a68c7](https://github.com/orhaniscoding/goconnect/commit/b4a68c729b0343b047a64f578fee0fbd3f83432c))
* **server:** implement network get/update/delete with idempotency and audit hooks ([#25](https://github.com/orhaniscoding/goconnect/issues/25)) ([b50ea58](https://github.com/orhaniscoding/goconnect/commit/b50ea58399d3de3c9ff258b95e8a44c3a32ad4bc))
* **server:** make rate limiter configurable via SERVER_RL_CAPACITY and SERVER_RL_WINDOW_MS ([#22](https://github.com/orhaniscoding/goconnect/issues/22)) ([a5938ad](https://github.com/orhaniscoding/goconnect/commit/a5938ad34073c560db2dde2cd869bc5891e25e16))

## 1.0.0 (2025-09-25)


### ⚠ BREAKING CHANGES

* **server:** new endpoints under /v1/networks/{id}/(join|approve|deny|kick|ban|members)

### Features

* add commitlint and husky for conventional commits ([#13](https://github.com/orhaniscoding/goconnect/issues/13)) ([cb87f62](https://github.com/orhaniscoding/goconnect/commit/cb87f6290594eaaa8f5a53f0ac2dff12c1c34736))
* **server:** implement Memberships & Join Flow v1 (domain, repo, service, handler, OpenAPI, docs) ([#15](https://github.com/orhaniscoding/goconnect/issues/15)) ([9e83e53](https://github.com/orhaniscoding/goconnect/commit/9e83e53f79c22cd9caa234d68a0e4ad3cfc9185d))
