# Changelog

## [1.1.0](https://github.com/orhaniscoding/goconnect/compare/v1.0.0...v1.1.0) (2025-09-29)


### Features

* **ipam:** add allocations listing endpoint and tests ([#35](https://github.com/orhaniscoding/goconnect/issues/35)) ([53acc79](https://github.com/orhaniscoding/goconnect/commit/53acc793dd5703fe0ac240345fa44255a13d973c))
* **ipam:** add base IPAM domain, repository, service, tests and spec placeholder ([#30](https://github.com/orhaniscoding/goconnect/issues/30)) ([b1c6fce](https://github.com/orhaniscoding/goconnect/commit/b1c6fce18361826a7e5c9a6c814fa33192d339af))
* **ipam:** add IP release endpoint, service logic, offset reuse and tests ([#37](https://github.com/orhaniscoding/goconnect/issues/37)) ([730b3db](https://github.com/orhaniscoding/goconnect/commit/730b3dbcfdcc0609f81d15c844a7841b6c1291bb))
* **ipam:** enforce membership for IP allocation and add tests ([#34](https://github.com/orhaniscoding/goconnect/issues/34)) ([b4a68c7](https://github.com/orhaniscoding/goconnect/commit/b4a68c729b0343b047a64f578fee0fbd3f83432c))
* **server:** implement network get/update/delete with idempotency and audit hooks ([#25](https://github.com/orhaniscoding/goconnect/issues/25)) ([b50ea58](https://github.com/orhaniscoding/goconnect/commit/b50ea58399d3de3c9ff258b95e8a44c3a32ad4bc))
* **server:** make rate limiter configurable via SERVER_RL_CAPACITY and SERVER_RL_WINDOW_MS ([#22](https://github.com/orhaniscoding/goconnect/issues/22)) ([a5938ad](https://github.com/orhaniscoding/goconnect/commit/a5938ad34073c560db2dde2cd869bc5891e25e16))

## 1.0.0 (2025-09-25)


### ⚠ BREAKING CHANGES

* **server:** new endpoints under /v1/networks/{id}/(join|approve|deny|kick|ban|members)

### Features

* add commitlint and husky for conventional commits ([#13](https://github.com/orhaniscoding/goconnect/issues/13)) ([cb87f62](https://github.com/orhaniscoding/goconnect/commit/cb87f6290594eaaa8f5a53f0ac2dff12c1c34736))
* **server:** implement Memberships & Join Flow v1 (domain, repo, service, handler, OpenAPI, docs) ([#15](https://github.com/orhaniscoding/goconnect/issues/15)) ([9e83e53](https://github.com/orhaniscoding/goconnect/commit/9e83e53f79c22cd9caa234d68a0e4ad3cfc9185d))
