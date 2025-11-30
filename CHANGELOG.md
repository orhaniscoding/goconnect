# Changelog

## [2.28.0](https://github.com/orhaniscoding/goconnect/compare/v2.27.0...v2.28.0) (2025-11-30)


### Features

* **server:** add interactive setup wizard with web UI ([#setup](https://github.com/orhaniscoding/goconnect/issues/setup))
* **daemon:** add interactive CLI setup command ([#daemon-setup](https://github.com/orhaniscoding/goconnect/issues/daemon-setup))
* **web-ui:** fix Next.js 15+ params async compatibility ([#webui](https://github.com/orhaniscoding/goconnect/issues/webui))


### Bug Fixes

* **migrations:** fix PostgreSQL schema for posts, devices, peers tables
* **migrations:** add proper up/down migration files for Goose format
* **server:** simplify tenant CREATE query for registration flow
* **web-ui:** fix locale params Promise handling in login/register pages


### Build

* **ci:** update release workflow to use GoReleaser v2
* **ci:** add .goreleaser.yaml for server and daemon


## [2.27.0](https://github.com/orhaniscoding/goconnect/compare/v2.26.0...v2.27.0) (2025-11-29)


### Features

* Complete GoConnect architecture cleanup and product-ready implementation ([abd9ad1](https://github.com/orhaniscoding/goconnect/commit/abd9ad1b76678e58df16bb76320f0ceee8616e81))
* **daemon,web:** implement localhost bridge, deep linking, and daemon discovery ([e22a2fb](https://github.com/orhaniscoding/goconnect/commit/e22a2fb6f28c85000b4e619e9ac8106254f5f6b9))
