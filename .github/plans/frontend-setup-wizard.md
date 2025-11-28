# Plan: Frontend Implementation - Zero-Config Setup Wizard

## 1. Executive Summary
This plan focuses on implementing the **Web UI** portion of the Zero-Config experience.
Currently, the backend supports Setup Mode logic (`/setup` endpoints), but the frontend lacks the interface.
We will create a modern, "Grandma-proof" installation wizard that guides the user through the initial server configuration without requiring any terminal interaction.

## 2. User Experience (UX) Flow

1.  **Detection:** When the user opens the app (e.g., `http://localhost:8080`), the backend redirects to `/setup` if no config exists.
2.  **Welcome Screen:** Friendly greeting. "Welcome to GoConnect. Let's get your private network running."
3.  **Step 1: Mode Selection:**
    *   **Personal (Recommended):** "I want to host a network for me and my friends." (Selects SQLite).
    *   **Enterprise:** "I want to connect to an external database." (Selects Postgres - toggle hidden/advanced by default).
4.  **Step 2: Admin Account:**
    *   Simple form: Email, Display Name, Password.
    *   No complex JWT/Secret questions (auto-generated in background).
5.  **Step 3: First Network (Optional but recommended):**
    *   "Name your first network" (e.g., "Minecraft Server", "Office LAN").
    *   Auto-generate CIDR in background.
6.  **Step 4: Finalize:**
    *   "Save & Start" button.
    *   **Crucial:** The UI must handle the backend restart gracefully. Show a "Configuring & Restarting..." spinner, then poll `/health` until the server is back up, then redirect to `/login`.

## 3. Technical Implementation (Next.js)

### 3.1. New Route: `/setup`
*   Create `web-ui/src/app/[locale]/setup/page.tsx`.
*   **Layout:** A simplified layout (no sidebar, no auth guards). Just a centered card/container.
*   **Middleware:** Ensure Next.js middleware allows access to `/setup` without a token.

### 3.2. Components (Shadcn UI)
*   **Stepper:** A visual indicator of progress (Step 1 of 3).
*   **Forms:** Use `react-hook-form` + `zod` for validation.
*   **Feedback:** Success/Error toasts using `sonner` or existing toast component.

### 3.3. State Management
*   Fetch setup status from `GET /setup/status` on mount.
*   If backend is NOT in setup mode, redirect to `/dashboard` (or `/login`).
*   Maintain wizard state (current step, form data).

### 3.4. The "Restart" Handover
This is the trickiest part.
1.  POST config to `/setup`.
2.  Backend saves `goconnect.yaml` and returns `restart_required: true`.
3.  Frontend shows "Restarting..." modal.
4.  Backend shuts down and restarts (managed by the OS service/binary loop or manual restart if dev).
5.  Frontend polls `GET /health` every 2 seconds.
6.  When `GET /health` returns 200, redirect to `/login`.

## 4. Tasks

- [ ] **Middleware Update:** Ensure `web-ui/src/middleware.ts` allows `/setup` route.
- [ ] **API Client:** Add `getSetupStatus`, `validateConfig`, and `persistConfig` to `web-ui/src/lib/api.ts`.
- [ ] **Setup Page Layout:** Create the focused layout for the wizard.
- [ ] **Wizard Logic:** Implement the multi-step form logic.
- [ ] **Restart Handler:** Implement the polling mechanism for server restart detection.
- [ ] **i18n:** Add English and Turkish translations for all setup strings.

## 5. Success Criteria
*   A user can download the empty binary, run it, go to `localhost:8080`, and complete the setup via UI.
*   After setup, the user is automatically redirected to login.
*   The `goconnect.yaml` file is created correctly on disk.
