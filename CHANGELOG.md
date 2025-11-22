# Changelog

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
