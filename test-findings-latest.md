# GoConnect Test Agent Findings (latest `main`)

This file documents the current test failures on the latest `main` (clean checkout after `git pull origin main`) and how to fix them. It is intended for the primary development agent to apply the fixes.

## Reproduction Commands
- Backend: `cd server && go test ./... -short`
- Daemon: `cd client-daemon && go test ./... -short`
- Frontend: `cd web-ui && npm run lint`

## Current Status (clean HEAD)
- Backend tests: pass.
- Daemon tests: pass.
- Frontend: `npm run lint` **fails** with multiple TypeScript/alias errors (details below).
- Race tests: blocked because gcc/CGO toolchain is not installed (`-race` requires C compiler).
- Build was not re-run on clean HEAD because lint already fails; expect the prior prerender error (`I18nProvider missing in tree` for `/admin`) to reappear once lint is fixed unless addressed.

## Frontend Lint Errors (clean HEAD)
From `npm run lint`:
1) `src/app/[locale]/setup/page.tsx`
   - TS2307 missing modules: `../lib/i18n-context`, `../components/ui/{button,input,label,card}`, `../lib/api`
   - TS7006 implicit `any` for event params at lines ~191, 201, 215
2) Admin pages/components
   - TS2307 missing modules: `@/lib/api`, `@/lib/auth`, `@/components/admin/{UserDetailDialog,EditRoleDialog,SuspendUserDialog}`
3) `src/components/Layout.tsx`
   - TS2614: `Navbar` is default export but imported as named.

## Likely Root Causes
- TS path aliases not configured: `@/*` not mapped, and the setup page uses relative paths expecting a different structure.
- UI primitives (`button/input/label/card`) are missing; imports fail.
- Layout import mismatch (default vs named).
- Event handlers lack explicit types under `strict`.
- Admin pages assume aliases that are unresolved without `baseUrl`/`paths`.
- Build-time I18nProvider issue: `/admin` renders without provider (observed in previous build attempt). Needs wrapping in provider at layout/page level.

## Fix Plan
1) Configure TS aliases:
   - In `web-ui/tsconfig.json` add:
     ```json
     "baseUrl": "./src",
     "paths": { "@/*": ["./*"] }
     ```
2) Create minimal UI components under `web-ui/src/components/ui/`:
   - `button.tsx`, `input.tsx`, `label.tsx`, `card.tsx` using simple classNames.
   - Add helper `web-ui/src/lib/ui.ts` with `cn(...)`.
3) Fix imports/typing in setup page:
   - Use alias imports: `@/lib/i18n-context`, `@/components/ui/*`, `@/lib/api`.
   - Add `use client` directive at top; type event handlers with `ChangeEvent<HTMLInputElement>`.
4) Fix Layout import:
   - `import Navbar from './Navbar'` and use default export.
5) Ensure admin pages resolve aliases:
   - After aliases set in tsconfig, adjust imports to use `@/...` consistently if needed.
6) Re-run `npm run lint`, then `npm run build`.
7) Address build prerender error:
   - `/admin` (and nested pages) need to be wrapped in `I18nProvider` (from `src/lib/i18n-context`) in their layout or page component so SSR/prerender has the provider.
8) Optional: Install gcc (e.g., mingw) to enable `go test -race` for server and daemon.

## Suggested Prompt for Primary Agent
```
You are on the latest main. Lint/build are failing in web-ui.
Tasks:
1) In web-ui/tsconfig.json add baseUrl './src' and paths mapping '@/*' -> './*'.
2) Add minimal UI primitives in web-ui/src/components/ui/{button,input,label,card}.tsx and helper web-ui/src/lib/ui.ts (with cn()).
3) Fix setup page at web-ui/src/app/[locale]/setup/page.tsx: add "use client", use alias imports (@/lib/i18n-context, @/components/ui/*, @/lib/api), type event handlers with ChangeEvent<HTMLInputElement>.
4) Fix Layout import in web-ui/src/components/Layout.tsx: import Navbar as default.
5) Ensure admin pages resolve alias imports (@/lib/api, @/lib/auth, admin dialogs).
6) Run `npm run lint` then `npm run build`; if build fails with "I18nProvider missing in tree" for /admin, wrap admin layout/page with I18nProvider from src/lib/i18n-context.
7) If possible, install gcc/enable CGO to run `go test ./... -race` for server and client-daemon; otherwise note as blocked.
```

## Notes
- A prior stash `codex-temp` contains the quick fixes implementing steps 1–4 and the UI components; apply via `git stash pop` if useful, or re-implement cleanly.
- All commands above assume you run from repo root.
