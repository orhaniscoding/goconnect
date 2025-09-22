# GoConnect — by orhaniscoding

© 2025 Orhan Tüzer (orhaniscoding) · MIT

**Binaries:** `goconnect-server`, `goconnect-daemon`  
**Web UI:** Next.js (TR/EN), unified user+admin.

## Quickstart
```bash
docker compose -f server/docker-compose.yaml up -d
go run ./server/cmd/server
go run ./client-daemon/cmd/daemon
cd web-ui && npm i && npm run dev
```
See `docs/TECH_SPEC.md` and `docs/SUPER_PROMPT.md`.
