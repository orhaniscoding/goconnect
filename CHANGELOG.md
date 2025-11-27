# Changelog

## [2.22.2](https://github.com/orhaniscoding/goconnect/compare/v2.22.1...v2.22.2) (2025-11-27)


### Bug Fixes

* **release:** complete release system audit and fixes ([35e8453](https://github.com/orhaniscoding/goconnect/commit/35e84534e58e25bccdd58bf304798e4a9a596ddb))

## [2.22.1](https://github.com/orhaniscoding/goconnect/compare/v2.22.0...v2.22.1) (2025-11-27)


### Bug Fixes

* **goreleaser:** fix YAML configuration errors ([336fd33](https://github.com/orhaniscoding/goconnect/commit/336fd33152d1c1ebf42576ac94ee3245a7247120))

## [2.22.0](https://github.com/orhaniscoding/goconnect/compare/v2.21.2...v2.22.0) (2025-11-27)


### Features

* **installer:** complete overhaul of installation system ([3d0668f](https://github.com/orhaniscoding/goconnect/commit/3d0668fa992661809908f81e2c4a29ac2396def0))

## [2.21.2](https://github.com/orhaniscoding/goconnect/compare/v2.21.1...v2.21.2) (2025-11-27)


### Bug Fixes

* **installer:** improve daemon installation experience ([e901c4d](https://github.com/orhaniscoding/goconnect/commit/e901c4d65f786fd718f1bde6e4bd518868fd2074))

## [2.21.1](https://github.com/orhaniscoding/goconnect/compare/v2.21.0...v2.21.1) (2025-11-27)


### Bug Fixes

* **handler:** correct error type assertion in admin SuspendUser handler ([86ec6da](https://github.com/orhaniscoding/goconnect/commit/86ec6da76827c8986a4e0fb6259d8c885e0733ce))

## [2.21.0](https://github.com/orhaniscoding/goconnect/compare/v2.20.0...v2.21.0) (2025-11-27)


### Features

* **auth:** implement real-time user suspension enforcement ([8cfb664](https://github.com/orhaniscoding/goconnect/commit/8cfb664f5ae9c175b562c77189707802bb213480))

## [2.20.0](https://github.com/orhaniscoding/goconnect/compare/v2.19.0...v2.20.0) (2025-11-27)


### Features

* **admin:** add comprehensive admin user management system ([9f36dba](https://github.com/orhaniscoding/goconnect/commit/9f36dba8f5c42b8d5bcddf9d143d49a816ad55ff))

## [2.19.0](https://github.com/orhaniscoding/goconnect/compare/v2.18.0...v2.19.0) (2025-11-27)


### Features

* **api:** add posts feature and user profile enhancement ([b0491a6](https://github.com/orhaniscoding/goconnect/commit/b0491a6f4d6ef374ad7ef27e18fd4109ccd970b5))
* **ui:** add VPN dashboard and social media feed UI ([3110cfc](https://github.com/orhaniscoding/goconnect/commit/3110cfc282c5fe35586c567069578592ef3d953f))

## [2.18.0](https://github.com/orhaniscoding/goconnect/compare/v2.17.0...v2.18.0) (2025-11-26)


### Features

* add peer service with handler and update gitignore ([#116](https://github.com/orhaniscoding/goconnect/issues/116)) ([03c1a53](https://github.com/orhaniscoding/goconnect/commit/03c1a53f30ae031b0d66a387c47a01f35ae1f02c))

## [2.17.0](https://github.com/orhaniscoding/goconnect/compare/v2.16.0...v2.17.0) (2025-11-26)


### Features

* **network:** add DNS/MTU/split_tunnel editing support ([071fd21](https://github.com/orhaniscoding/goconnect/commit/071fd21c088a1bedf451566ddf95c991e7a93c20))
* **web-ui:** add edit indicator to Network and Tenant chat messages ([71df624](https://github.com/orhaniscoding/goconnect/commit/71df624c3801214e4ee957019ed462200d4d4bf7))
* **web-ui:** add online status indicators to tenant members page ([d4d1a97](https://github.com/orhaniscoding/goconnect/commit/d4d1a97af77dab2f7ad109f966186d4047eb0837))
* **web-ui:** add UI components and dashboard improvements ([0e62d7d](https://github.com/orhaniscoding/goconnect/commit/0e62d7d2f6768c51a71a9e8ac35eafb332340e7c))


### Bug Fixes

* **auth:** enforce OIDC state validation ([c7fbd76](https://github.com/orhaniscoding/goconnect/commit/c7fbd768cfc7a776a7c41d5faa40ccbddb18381b))

## [2.16.0](https://github.com/orhaniscoding/goconnect/compare/v2.15.0...v2.16.0) (2025-11-26)


### Features

* **web-ui:** add online status indicators to network chat ([4169d07](https://github.com/orhaniscoding/goconnect/commit/4169d07d728d7edeef3987869776d187fb8f8fa4))

## [2.15.0](https://github.com/orhaniscoding/goconnect/compare/v2.14.0...v2.15.0) (2025-11-26)


### Features

* **web-ui:** add file attachment support to network chat ([c4c65fb](https://github.com/orhaniscoding/goconnect/commit/c4c65fb899e9f9f49edd169893df8c14df7f5358))
* **web-ui:** add global navigation bar component ([1294584](https://github.com/orhaniscoding/goconnect/commit/1294584ce66bc201e389618b4655d36ae4206b75))

## [2.14.0](https://github.com/orhaniscoding/goconnect/compare/v2.13.0...v2.14.0) (2025-11-26)


### Features

* **web-ui:** add network chat page with real-time messaging ([e5bf969](https://github.com/orhaniscoding/goconnect/commit/e5bf96933c590cd2991c4af02da0fc5857936e79))
* **web-ui:** enhance audit log page with filters and color-coded badges ([2b8109d](https://github.com/orhaniscoding/goconnect/commit/2b8109d755ec613bad9ffc6362798a8b3d443dd8))

## [2.13.0](https://github.com/orhaniscoding/goconnect/compare/v2.12.0...v2.13.0) (2025-11-26)


### Features

* **server,ui:** add tenant member ban functionality ([62a4c30](https://github.com/orhaniscoding/goconnect/commit/62a4c307d2c2d8e8205bc8dfae4e063b0d796250))
* **server:** add DELETE /v1/tenants/{id} endpoint for tenant deletion ([148badd](https://github.com/orhaniscoding/goconnect/commit/148badd2f4f5d6a7e09d0a3bbfc3697d8616a2f8))
* **tenant:** add unban member and list banned members functionality ([5c9aeef](https://github.com/orhaniscoding/goconnect/commit/5c9aeefb1a435eb5ad7d127c02a28ea9258e73da))
* **ui:** implement tenant deletion with confirmation modal ([c3f0762](https://github.com/orhaniscoding/goconnect/commit/c3f07627d7c60b4e62bfadf6f8d40a2412ed1578))
* **web-ui:** add banned members tab to tenant settings page ([0d79a1c](https://github.com/orhaniscoding/goconnect/commit/0d79a1c2ede2738e1cbf05167ff382855c7e01cb))

## [2.12.0](https://github.com/orhaniscoding/goconnect/compare/v2.11.0...v2.12.0) (2025-11-25)


### Features

* **api:** implement PATCH /v1/tenants/{id} for tenant settings update ([d67167f](https://github.com/orhaniscoding/goconnect/commit/d67167fe1f825270ace209d391df041dc8dfa850))
* **ui:** add Tenant Settings page for owner/admin ([4ca1090](https://github.com/orhaniscoding/goconnect/commit/4ca10905d579f76409f6f7624fad53d82c83cf51))

## [2.11.0](https://github.com/orhaniscoding/goconnect/compare/v2.10.0...v2.11.0) (2025-11-25)


### Features

* implement real-time WebSocket tenant chat with typing indicators ([3b01f76](https://github.com/orhaniscoding/goconnect/commit/3b01f763dbc9a407acd62b173442a54aef9932fa))

## [2.10.0](https://github.com/orhaniscoding/goconnect/compare/v2.9.1...v2.10.0) (2025-11-25)


### Features

* **server:** integrate tenant multi-membership system into main.go ([1467ea5](https://github.com/orhaniscoding/goconnect/commit/1467ea5b9fac783066b98e58f8c54ff2e3f66ca4))

## [2.9.1](https://github.com/orhaniscoding/goconnect/compare/v2.9.0...v2.9.1) (2025-11-25)


### Bug Fixes

* **lint:** resolve client-daemon linting errors ([635d745](https://github.com/orhaniscoding/goconnect/commit/635d7451436c082a10c4b8e4d3fb8d3fb55d5bf8))

## [2.9.0](https://github.com/orhaniscoding/goconnect/compare/v2.8.8...v2.9.0) (2025-11-25)


### Features

* **api:** add HTTP handlers for tenant multi-membership system ([b984827](https://github.com/orhaniscoding/goconnect/commit/b984827a86fb1c3eb90688510f84eb80313fe407))
* **auth:** implement 2FA recovery codes for account recovery ([cab830f](https://github.com/orhaniscoding/goconnect/commit/cab830ff472dcf7fcbeb20510017933be5b697ea))
* **dashboard:** add tenants quick action link ([9575953](https://github.com/orhaniscoding/goconnect/commit/9575953e95d959763c2902ae24b17c4a54e39251))
* **repo:** add PostgreSQL implementations for InviteToken and IPRule ([cd3db72](https://github.com/orhaniscoding/goconnect/commit/cd3db72374e0d53df92cd0b9e2c9e8768cbfc6d4))
* **tenant:** implement multi-membership system foundation ([3deedb2](https://github.com/orhaniscoding/goconnect/commit/3deedb27f4305eca6352aff290c05b7918e9b15e))
* **web-ui:** add tenant API client and i18n translations ([ae54564](https://github.com/orhaniscoding/goconnect/commit/ae545646250fdc56f0857d9c7781ed1f22db238b))
* **web-ui:** add tenant pages (discover, detail, chat) ([5ec1661](https://github.com/orhaniscoding/goconnect/commit/5ec1661d32a9bc8e1409670efbfbf51c88ff8b23))
* **web:** add Recovery Codes management UI to Settings page ([3c51fbb](https://github.com/orhaniscoding/goconnect/commit/3c51fbb2bb14ff9f84e3c1e97e99296a14655cf9))


### Bug Fixes

* add KendimeNotlarim.md to .gitignore ([b0e70e2](https://github.com/orhaniscoding/goconnect/commit/b0e70e26c2c21f4fedf1c351ecf32ac0d4f5c84b))
* ensure services restart automatically on all platforms ([60542f5](https://github.com/orhaniscoding/goconnect/commit/60542f514d563fcd03d4cf20a3be2420d14566c6))

## [2.8.8](https://github.com/orhaniscoding/goconnect/compare/v2.8.7...v2.8.8) (2025-11-25)


### Bug Fixes

* Docker build - add .gitkeep to public folder ([77390fb](https://github.com/orhaniscoding/goconnect/commit/77390fb9d0a3fedd09e736f3617932b9f7ef6d32))

## [2.8.7](https://github.com/orhaniscoding/goconnect/compare/v2.8.6...v2.8.7) (2025-11-25)


### Bug Fixes

* add standalone output for Next.js Docker build ([35460ea](https://github.com/orhaniscoding/goconnect/commit/35460eae6f23faf3554ca6722ecf8a64ef4d318a))

## [2.8.6](https://github.com/orhaniscoding/goconnect/compare/v2.8.5...v2.8.6) (2025-11-25)


### Bug Fixes

* data race in TestHub_HandleInboundMessage ([e1b61e6](https://github.com/orhaniscoding/goconnect/commit/e1b61e68ace45b618134cf892f9caf90eac85742))
* golangci-lint errors and code formatting ([a422112](https://github.com/orhaniscoding/goconnect/commit/a422112848de0090e6feb0057edefde93c83ec32))
* TypeScript errors in i18n and API calls, upgrade golangci-lint to v1.64.8 ([cdf3362](https://github.com/orhaniscoding/goconnect/commit/cdf3362f500fc576bb4f796d4f73ebd297b4468f))

## [2.8.5](https://github.com/orhaniscoding/goconnect/compare/v2.8.4...v2.8.5) (2025-11-25)


### Bug Fixes

* **release:** optimize assets, fix Docker builds, improve release notes ([b87112e](https://github.com/orhaniscoding/goconnect/commit/b87112e90a7bbaa58bf7b84ed960d57c78551415))

## [2.8.4](https://github.com/orhaniscoding/goconnect/compare/v2.8.3...v2.8.4) (2025-11-25)


### Bug Fixes

* **release:** add Syft installation for SBOM generation ([7fb53f2](https://github.com/orhaniscoding/goconnect/commit/7fb53f2a136b9a6eca9d68d9e4808f5467fb80c9))

## [2.8.3](https://github.com/orhaniscoding/goconnect/compare/v2.8.2...v2.8.3) (2025-11-25)


### Bug Fixes

* **release:** use external script files for nfpm package scripts ([589e7d9](https://github.com/orhaniscoding/goconnect/commit/589e7d95bc45144cbb768b34b0c71e51a5d80c06))

## [2.8.2](https://github.com/orhaniscoding/goconnect/compare/v2.8.1...v2.8.2) (2025-11-25)


### Bug Fixes

* **release:** add windows/arm64 ignore to daemon build for archive consistency ([063f245](https://github.com/orhaniscoding/goconnect/commit/063f245286295020796e957ee91c882366edac54))

## [2.8.1](https://github.com/orhaniscoding/goconnect/compare/v2.8.0...v2.8.1) (2025-11-25)


### Bug Fixes

* **release:** chain GoReleaser after Release Please for proper binary assets ([ea7cc37](https://github.com/orhaniscoding/goconnect/commit/ea7cc37e416ca99fc05e94fcd95f67126ea33b28))

## [2.8.0](https://github.com/orhaniscoding/goconnect/compare/v2.7.0...v2.8.0) (2025-11-25)


### Features

* **audit:** add filtering support to audit log queries ([adb597d](https://github.com/orhaniscoding/goconnect/commit/adb597d9fd8a3c4e23092e29cbfb21218e5ceb02))

## [2.7.0](https://github.com/orhaniscoding/goconnect/compare/v2.6.0...v2.7.0) (2025-11-25)


### Features

* **iprules:** implement admin-configurable IP allowlist/denylist per tenant ([e10adf6](https://github.com/orhaniscoding/goconnect/commit/e10adf64a04f97e4ec57f5fed39e057ef2247e94))

## [2.6.0](https://github.com/orhaniscoding/goconnect/compare/v2.5.0...v2.6.0) (2025-11-25)


### Features

* **gdpr:** implement GDPR/DSR endpoints for data export and deletion ([f370735](https://github.com/orhaniscoding/goconnect/commit/f37073564d30dcb612f9c4bd3be60e7670607ba3))
* **invite:** implement network invite token system ([dfb8c51](https://github.com/orhaniscoding/goconnect/commit/dfb8c51dd0c7110aa44d645e229f0dd724178ab0))
* **ratelimit:** implement endpoint-specific rate limiting ([5e1608f](https://github.com/orhaniscoding/goconnect/commit/5e1608fbd7291afa0a7eabcd4cad501c3a8947ba))

## [2.5.0](https://github.com/orhaniscoding/goconnect/compare/v2.4.1...v2.5.0) (2025-11-25)


### Features

* comprehensive CI/CD with Docker, multi-platform builds, and installation guides ([ea46937](https://github.com/orhaniscoding/goconnect/commit/ea46937651726f430449bf95c19ccb60159dc16a))

## [2.4.1](https://github.com/orhaniscoding/goconnect/compare/v2.4.0...v2.4.1) (2025-11-25)


### Bug Fixes

* **audit:** correct flaky age retention tamper detection test ([0b2f90b](https://github.com/orhaniscoding/goconnect/commit/0b2f90ba3f4b05d4bbcc318d8eb36acb158b241d))

## [2.4.0](https://github.com/orhaniscoding/goconnect/compare/v2.3.0...v2.4.0) (2025-11-25)


### Features

* add wireguard metrics (peers, bytes, handshake) to prometheus endpoint ([53dc11b](https://github.com/orhaniscoding/goconnect/commit/53dc11bf27c898941d11d3759ad11e3727505bf8))
* **admin:** add pagination controls to all tabs ([43b7368](https://github.com/orhaniscoding/goconnect/commit/43b73683437e40935faa458efe9d483310a0a36a))
* **admin:** complete device management with delete functionality ([263e3b5](https://github.com/orhaniscoding/goconnect/commit/263e3b5fe53548a1163c3c4e08fa6d4d21df62d8))
* **auth:** implement OIDC user provisioning (JIT) ([b50e04b](https://github.com/orhaniscoding/goconnect/commit/b50e04b290e82092e5baa902f1e69c39edaa0a77))
* **auth:** implement token blacklist with Redis ([4333141](https://github.com/orhaniscoding/goconnect/commit/4333141cbd8c1d737a69325295a72246549efbc6))
* **daemon:** improve OS detection and Linux DNS configuration ([efdc5df](https://github.com/orhaniscoding/goconnect/commit/efdc5dfd6d1bcfee52579864d2014c7b3bd0ca35))
* **device:** implement offline detection worker ([648338e](https://github.com/orhaniscoding/goconnect/commit/648338ef6498a644b00e75a7aaf4a1bacbb0508c))
* **device:** implement offline detection worker and fix audit logs ([8363d23](https://github.com/orhaniscoding/goconnect/commit/8363d232c20a7a7f13b9988211f6590047fb2710))
* implement file upload progress signaling ([0ef96c4](https://github.com/orhaniscoding/goconnect/commit/0ef96c4a7a048b27df7959ff67d5cdbc74d8d96c))
* implement message threading, typing indicators, and screen sharing signaling ([b9307d8](https://github.com/orhaniscoding/goconnect/commit/b9307d82d4e195965257e91d5cdd1bb3e0cfe19b))
* implement OIDC (SSO) backend support ([ecc335c](https://github.com/orhaniscoding/goconnect/commit/ecc335c2fb76c60b88cfbefed094f05dda628264))
* implement search and filtering for admin dashboard resources ([133751b](https://github.com/orhaniscoding/goconnect/commit/133751bec3a2d33a17475a198226f371b2fad562))
* implement search and filtering in admin dashboard ([2581df4](https://github.com/orhaniscoding/goconnect/commit/2581df45435b137070ae7544ba6f78b3807aca1e))
* implement websocket rate limiting ([c50956f](https://github.com/orhaniscoding/goconnect/commit/c50956f88b46c9c067beb07bef68c91233a9a3a4))
* **server:** add graceful shutdown for HTTP server and background workers ([218595a](https://github.com/orhaniscoding/goconnect/commit/218595ad6a9fe975d3edf9ed4711957083e9b1b2))
* **server:** populate DNS servers in WireGuard profile JSON response ([91e75a8](https://github.com/orhaniscoding/goconnect/commit/91e75a82cc5cc5e0c78490c461fbe929a5ce15c8))
* **ui:** add OIDC login button and callback handler ([bf413d1](https://github.com/orhaniscoding/goconnect/commit/bf413d1ef77cd43080b870b0cfefe5b05ee96201))
* **ui:** add settings page with 2FA management ([7640463](https://github.com/orhaniscoding/goconnect/commit/7640463fe86e330645b8cc64c1b3e2cac5a8d233))
* **ui:** add settings page with 2FA management ([19df189](https://github.com/orhaniscoding/goconnect/commit/19df1895e1ef4352735d2f29cc9afcde5abc9b7c))
* **ui:** display auth provider in admin user list ([52ab306](https://github.com/orhaniscoding/goconnect/commit/52ab306935a84d25d26a3b8187b153c608c72e56))
* **websocket:** implement device online/offline events ([d778406](https://github.com/orhaniscoding/goconnect/commit/d778406d413988834ae28b6bb9801239b86ea106))
* **websocket:** implement direct messages ([ab48094](https://github.com/orhaniscoding/goconnect/commit/ab4809422d72b11e14ce1d4c9739705b8aa4349f))
* **websocket:** implement message reactions ([4a86f9b](https://github.com/orhaniscoding/goconnect/commit/4a86f9bacce7c4cda7f55ce4b02c33a8a30fba34))
* **websocket:** implement message read receipts ([ac5461b](https://github.com/orhaniscoding/goconnect/commit/ac5461b5005d2b5886a32c227c6a5858ffd03ac8))
* **websocket:** implement network membership events broadcast ([a31ac4f](https://github.com/orhaniscoding/goconnect/commit/a31ac4f84c4817806d492850dd4d92ad82bb5d73))
* **websocket:** implement presence status broadcast ([eb9b124](https://github.com/orhaniscoding/goconnect/commit/eb9b124da78073a701915022f6dd3a22d56c8753))
* **websocket:** implement rate limiting and room authorization ([77e6f33](https://github.com/orhaniscoding/goconnect/commit/77e6f33c9c253df07ea0a7bf27b39add91f44e02))
* **websocket:** implement voice/video call signaling ([f333228](https://github.com/orhaniscoding/goconnect/commit/f333228c98b230086b929be2a9a082190a49e4eb))


### Bug Fixes

* **daemon:** implement missing configurator methods for darwin ([3509aa8](https://github.com/orhaniscoding/goconnect/commit/3509aa820e7821c2b055c47d4765c627162022dc))
* revert manifest to align with Release Please ([f930bc2](https://github.com/orhaniscoding/goconnect/commit/f930bc2553d423842c7df550965d219d2ea741da))

## [2.4.0](https://github.com/orhaniscoding/goconnect/compare/v2.3.0...v2.4.0) (2025-11-25)


### Features

* add wireguard metrics (peers, bytes, handshake) to prometheus endpoint ([53dc11b](https://github.com/orhaniscoding/goconnect/commit/53dc11bf27c898941d11d3759ad11e3727505bf8))
* **admin:** add pagination controls to all tabs ([43b7368](https://github.com/orhaniscoding/goconnect/commit/43b73683437e40935faa458efe9d483310a0a36a))
* **admin:** complete device management with delete functionality ([263e3b5](https://github.com/orhaniscoding/goconnect/commit/263e3b5fe53548a1163c3c4e08fa6d4d21df62d8))
* **auth:** implement OIDC user provisioning (JIT) ([b50e04b](https://github.com/orhaniscoding/goconnect/commit/b50e04b290e82092e5baa902f1e69c39edaa0a77))
* **auth:** implement token blacklist with Redis ([4333141](https://github.com/orhaniscoding/goconnect/commit/4333141cbd8c1d737a69325295a72246549efbc6))
* **daemon:** improve OS detection and Linux DNS configuration ([efdc5df](https://github.com/orhaniscoding/goconnect/commit/efdc5dfd6d1bcfee52579864d2014c7b3bd0ca35))
* **device:** implement offline detection worker ([648338e](https://github.com/orhaniscoding/goconnect/commit/648338ef6498a644b00e75a7aaf4a1bacbb0508c))
* **device:** implement offline detection worker and fix audit logs ([8363d23](https://github.com/orhaniscoding/goconnect/commit/8363d232c20a7a7f13b9988211f6590047fb2710))
* implement file upload progress signaling ([0ef96c4](https://github.com/orhaniscoding/goconnect/commit/0ef96c4a7a048b27df7959ff67d5cdbc74d8d96c))
* implement message threading, typing indicators, and screen sharing signaling ([b9307d8](https://github.com/orhaniscoding/goconnect/commit/b9307d82d4e195965257e91d5cdd1bb3e0cfe19b))
* implement OIDC (SSO) backend support ([ecc335c](https://github.com/orhaniscoding/goconnect/commit/ecc335c2fb76c60b88cfbefed094f05dda628264))
* implement search and filtering for admin dashboard resources ([133751b](https://github.com/orhaniscoding/goconnect/commit/133751bec3a2d33a17475a198226f371b2fad562))
* implement search and filtering in admin dashboard ([2581df4](https://github.com/orhaniscoding/goconnect/commit/2581df45435b137070ae7544ba6f78b3807aca1e))
* implement websocket rate limiting ([c50956f](https://github.com/orhaniscoding/goconnect/commit/c50956f88b46c9c067beb07bef68c91233a9a3a4))
* **server:** add graceful shutdown for HTTP server and background workers ([218595a](https://github.com/orhaniscoding/goconnect/commit/218595ad6a9fe975d3edf9ed4711957083e9b1b2))
* **server:** populate DNS servers in WireGuard profile JSON response ([91e75a8](https://github.com/orhaniscoding/goconnect/commit/91e75a82cc5cc5e0c78490c461fbe929a5ce15c8))
* **ui:** add OIDC login button and callback handler ([bf413d1](https://github.com/orhaniscoding/goconnect/commit/bf413d1ef77cd43080b870b0cfefe5b05ee96201))
* **ui:** add settings page with 2FA management ([7640463](https://github.com/orhaniscoding/goconnect/commit/7640463fe86e330645b8cc64c1b3e2cac5a8d233))
* **ui:** add settings page with 2FA management ([19df189](https://github.com/orhaniscoding/goconnect/commit/19df1895e1ef4352735d2f29cc9afcde5abc9b7c))
* **ui:** display auth provider in admin user list ([52ab306](https://github.com/orhaniscoding/goconnect/commit/52ab306935a84d25d26a3b8187b153c608c72e56))
* **websocket:** implement device online/offline events ([d778406](https://github.com/orhaniscoding/goconnect/commit/d778406d413988834ae28b6bb9801239b86ea106))
* **websocket:** implement direct messages ([ab48094](https://github.com/orhaniscoding/goconnect/commit/ab4809422d72b11e14ce1d4c9739705b8aa4349f))
* **websocket:** implement message reactions ([4a86f9b](https://github.com/orhaniscoding/goconnect/commit/4a86f9bacce7c4cda7f55ce4b02c33a8a30fba34))
* **websocket:** implement message read receipts ([ac5461b](https://github.com/orhaniscoding/goconnect/commit/ac5461b5005d2b5886a32c227c6a5858ffd03ac8))
* **websocket:** implement network membership events broadcast ([a31ac4f](https://github.com/orhaniscoding/goconnect/commit/a31ac4f84c4817806d492850dd4d92ad82bb5d73))
* **websocket:** implement presence status broadcast ([eb9b124](https://github.com/orhaniscoding/goconnect/commit/eb9b124da78073a701915022f6dd3a22d56c8753))
* **websocket:** implement rate limiting and room authorization ([77e6f33](https://github.com/orhaniscoding/goconnect/commit/77e6f33c9c253df07ea0a7bf27b39add91f44e02))
* **websocket:** implement voice/video call signaling ([f333228](https://github.com/orhaniscoding/goconnect/commit/f333228c98b230086b929be2a9a082190a49e4eb))


### Bug Fixes

* **daemon:** implement missing configurator methods for darwin ([3509aa8](https://github.com/orhaniscoding/goconnect/commit/3509aa820e7821c2b055c47d4765c627162022dc))

## [Unreleased]

### Features

## [2.8.0] - 2025-11-25

### Features

* **server:** add signal-aware graceful shutdown for HTTP server, background workers, and audit pipelines
* **server:** populate DNS servers from network config in WireGuard profile JSON response

## [2.7.0] - 2025-11-25

### Features

* **auth:** implement token blacklist with Redis
* **daemon:** improve OS detection and Linux DNS configuration
* **ui:** refactor login and register pages to use i18n

## [2.6.0] - 2025-11-25

### Features

* **auth:** implement OIDC (SSO) backend with JIT user provisioning
* **auth:** implement OIDC login flow in frontend
* **ui:** add settings page with 2FA management (enable/disable)
* **ui:** add "Login with SSO" button
* **ui:** display auth provider in admin user list

## [2.5.0] - 2025-11-23

### Features

* **metrics:** add WireGuard interface metrics (peers, bytes, handshake) to Prometheus endpoint
* **websocket:** implement rate limiting (10 msg/s, burst 20) per client
* **websocket:** implement direct messages (DMs) with scope canonicalization
* **websocket:** implement read receipts (`chat.read`)
* **websocket:** implement call signaling for WebRTC (audio/video/screen)
* **websocket:** implement message reactions (`chat.reaction`)
* **websocket:** implement message threads (reply support)
* **websocket:** implement typing indicators (`chat.typing`)
* **websocket:** implement file upload progress signaling (`file.upload`)

## [2.4.0] (2025-11-22)


### Features

* implement search and filtering for users, tenants, networks, and devices in admin dashboard

## [2.3.0](https://github.com/orhaniscoding/goconnect/compare/v2.2.0...v2.3.0) (2025-11-21)


### Features

* implement network management in admin dashboard ([8ceb912](https://github.com/orhaniscoding/goconnect/commit/8ceb9123182b70921ea5d657720c55b722087ea7))
* implement device management in admin dashboard ([263e3b5](https://github.com/orhaniscoding/goconnect/commit/263e3b5))

## [2.2.0](https://github.com/orhaniscoding/goconnect/compare/v2.1.0...v2.2.0) (2025-11-21)


### Features

* implement delete user functionality ([0e95607](https://github.com/orhaniscoding/goconnect/commit/0e95607b295fd1a8de53b5ee9b55b0d2f73d4aa5))

## [2.1.0](https://github.com/orhaniscoding/goconnect/compare/v2.0.0...v2.1.0) (2025-11-21)


### Features

* Add admin panel with system management ([95ea8d0](https://github.com/orhaniscoding/goconnect/commit/95ea8d08c71ceae494e9d21c6c9f58c6b6fd80c5))
* add audit logs and wireguard config generation ([0bfd1d4](https://github.com/orhaniscoding/goconnect/commit/0bfd1d4dcbf68d0345bf4c0fad00d9e435e593a3))
* add audit logs tab to admin dashboard ([046e51a](https://github.com/orhaniscoding/goconnect/commit/046e51a72348e69e392a83fef306c43a84b0d403))
* add connect/disconnect control to client daemon and dashboard ([84f81bd](https://github.com/orhaniscoding/goconnect/commit/84f81bdd5563b6cf0f7bcf2c6f178ddbcaa970e9))
* add device config endpoint for client daemon ([cf724ce](https://github.com/orhaniscoding/goconnect/commit/cf724ce160bd2540373660d0cb8437a41f3377c2))
* Add global notification/toast system ([5865922](https://github.com/orhaniscoding/goconnect/commit/5865922ba6408ed68b3b2906143befb9f89da097))
* Add message edit and delete UI ([5d2d8b2](https://github.com/orhaniscoding/goconnect/commit/5d2d8b2eca0dac6050ee0a3e6ceaedd34c08c3f1))
* Add message moderation UI for moderators ([f9ea9f9](https://github.com/orhaniscoding/goconnect/commit/f9ea9f9cc42f4e3fd588eaf8185751733c9ebaf9))
* Add Network-Scoped Chat with Room Management ([712e6e6](https://github.com/orhaniscoding/goconnect/commit/712e6e6a3daa3604b3c37c3b31b83d7302779811))
* Add one-click device registration in Web UI ([2ada428](https://github.com/orhaniscoding/goconnect/commit/2ada428d8ac55d33f7eadab04acfafbd8cbc146d))
* add online status indicator to network members list ([9fd81b0](https://github.com/orhaniscoding/goconnect/commit/9fd81b038998950ff5d3eebdfb9ffc7cfd63ff10))
* add online users list and presence updates to chat ([b8f52c4](https://github.com/orhaniscoding/goconnect/commit/b8f52c40d7985a6a329996f7e5a35034649fc6b0))
* add robust macos installer script and update makefile ([221bca5](https://github.com/orhaniscoding/goconnect/commit/221bca5ff10a82060f5170fc5960f81a44c2ba7a))
* Add user profile and settings page ([bc6ef1f](https://github.com/orhaniscoding/goconnect/commit/bc6ef1f690a80e16867c4cfd124a306f9e4c0fa4))
* **auth:** implement TOTP-based 2FA ([ae068e6](https://github.com/orhaniscoding/goconnect/commit/ae068e618d498f0b248cb4521d53274e281ffdf7))
* **auth:** implement TOTP-based 2FA ([baf123b](https://github.com/orhaniscoding/goconnect/commit/baf123bf4c16ddc7e7eb500c0562e44822982b2d))
* **client:** implement MagicDNS Lite (hosts file management) ([dada230](https://github.com/orhaniscoding/goconnect/commit/dada2309254a2a2337da34e16885bc89d4b0701e))
* **daemon:** add robust linux install script and update makefile ([7515f84](https://github.com/orhaniscoding/goconnect/commit/7515f8456e48124abed49cc9cd17a925131d0617))
* **daemon:** implement heartbeat engine ([63eb6f4](https://github.com/orhaniscoding/goconnect/commit/63eb6f4e9e8ec0f8b2caacef7b021cce81208c39))
* **daemon:** implement OS-specific network configuration ([721b7af](https://github.com/orhaniscoding/goconnect/commit/721b7af47c7bb57130569330dec47b42ac04ccc2))
* **daemon:** implement real-time wireguard status monitoring ([04c2016](https://github.com/orhaniscoding/goconnect/commit/04c20163bbe9fb499afee2175b88fd31fe25cd7b))
* **daemon:** implement routing and interface creation ([64161c3](https://github.com/orhaniscoding/goconnect/commit/64161c3be65b7d42f5e53f9c5adc76b10561ce53))
* **daemon:** implement wireguard interface configuration ([ac4a37d](https://github.com/orhaniscoding/goconnect/commit/ac4a37d49085f13b6e4884b9cae133cb2f8abad6))
* **daemon:** make interface name configurable via env var ([1c5a5de](https://github.com/orhaniscoding/goconnect/commit/1c5a5debb59ca3d26ee7f28c3ece6117e5a28b06))
* enforce network membership check for chat room join ([3d9c906](https://github.com/orhaniscoding/goconnect/commit/3d9c90677931dfd641636f2f4db8a51a6e0cef8f))
* implement active connections and messages today stats for admin dashboard ([d5fcd9b](https://github.com/orhaniscoding/goconnect/commit/d5fcd9b33c83bc2b401c4cf3536ebacc0ee6ec35))
* implement admin dashboard backend and frontend integration ([267a631](https://github.com/orhaniscoding/goconnect/commit/267a631c151f5d39637d70b3a05a8bb483c93221))
* implement audit log listing and admin ui ([bd00c03](https://github.com/orhaniscoding/goconnect/commit/bd00c03ebe701420e9b68a862c7fd2495a25bb16))
* implement chat attachments and password change functionality ([162ea43](https://github.com/orhaniscoding/goconnect/commit/162ea43441171095d48c2c818307c4a3c7ced2a7))
* Implement Client Daemon identity and registration ([2bb3303](https://github.com/orhaniscoding/goconnect/commit/2bb33035ca5c4d13cd8744b0ea21d68b4ad93fa3))
* implement count methods for network and device repositories to populate admin stats ([69ea2a9](https://github.com/orhaniscoding/goconnect/commit/69ea2a98c9eeed5d75f45fc3eb9e4548e511fe04))
* implement delete tenant functionality ([dabcb97](https://github.com/orhaniscoding/goconnect/commit/dabcb974d611c685832b2264fe03d050aefd2875))
* implement device config sync and enable endpoint ([92b12af](https://github.com/orhaniscoding/goconnect/commit/92b12af34e624df6a3538d0620e7d65f34309c5b))
* implement file attachments for chat ([3bae214](https://github.com/orhaniscoding/goconnect/commit/3bae21400421c947239b5f043af9971556c4250e))
* implement magicdns, monitoring, chat backend, 2fa backend and update docs ([77ea547](https://github.com/orhaniscoding/goconnect/commit/77ea5476dcad066e358c9b4caae51f49ab8da0d3))
* implement server-side wireguard interface management and peer sync ([47b01b3](https://github.com/orhaniscoding/goconnect/commit/47b01b35263791546d95ea964c7da6798471fffc))
* implement toggle admin status functionality in admin dashboard ([f69eba1](https://github.com/orhaniscoding/goconnect/commit/f69eba177caab3e78f165fff767173eed9037ff5))
* implement toggle admin status functionality in admin dashboard ([d8397cd](https://github.com/orhaniscoding/goconnect/commit/d8397cdb835bf2f77a51f01ff8da28d337814a86))
* implement websocket token refresh and standardize qr code generation ([5f6be7d](https://github.com/orhaniscoding/goconnect/commit/5f6be7d99672f00fe462e6876281287d5a52cc3b))
* Integrate notification system into Login and Profile pages ([4f58973](https://github.com/orhaniscoding/goconnect/commit/4f589735d5aac98e768a529042d27d47eed29759))
* integrate web-ui with client-daemon (fixed port 12345, status polling) ([4bf4ad6](https://github.com/orhaniscoding/goconnect/commit/4bf4ad6ae512d20348e611a3a593d25141928ddb))
* **server:** expose peer hostnames in wireguard config and API ([4f0ad02](https://github.com/orhaniscoding/goconnect/commit/4f0ad02b5cba9a49bf0056d266055724686b1fca))
* **ui:** add device registration card to dashboard ([9a16ec9](https://github.com/orhaniscoding/goconnect/commit/9a16ec926b5cd3de3bfc048aaaed7fab2f333b2c))
* **ui:** display device IP in dashboard and improve windows installer ([f6697f1](https://github.com/orhaniscoding/goconnect/commit/f6697f1e173fb07f5716c38bf81044d54a2e5242))
* **web:** implement chat history fetching ([7701176](https://github.com/orhaniscoding/goconnect/commit/770117674ac142778e39f6fba2d0f4c423d34da7))


### Bug Fixes

* formatting ([cf70480](https://github.com/orhaniscoding/goconnect/commit/cf70480dc37546140c3deb03d8cfb8793134bf2d))

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

* **repo:** enforce tenant isolation in postgres network repository
* **tests:** resolve all handler test failures and parameter order bugs

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

## [2.1.0](https://github.com/orhaniscoding/goconnect/compare/v2.0.0...v2.1.0) (2025-11-20)


### Features

* **admin:** implement network management in admin dashboard
* **admin:** implement delete user functionality
* **admin:** implement delete tenant functionality
* **admin:** optimize system stats with efficient count queries
* **tests:** fix regression in backend test suite
