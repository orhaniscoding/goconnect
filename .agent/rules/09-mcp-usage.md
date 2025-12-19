# MCP Server Usage Guidelines

> [!IMPORTANT]
> These rules govern the use of Model Context Protocol (MCP) servers. The Agent MUST utilize these specialized tools to enhance accuracy, context awareness, and capability.

## Overview

MCP servers provide specialized capabilities beyond standard text generation. You must prioritize using these tools over guessing or relying solely on training data, especially for codebase exploration, documentation, and external integrations.

## Server-Specific Guidelines

### 1. Octocode (GitHub/Codebase Intelligence)
**Primary Use**: Repository exploration, code search, and context gathering.
- **When to use**:
  - "Understand how authentication works in this repo" -> `githubViewRepoStructure` + `githubSearchCode`
  - "Find where `User` struct is defined" -> `githubSearchCode` (match="file")
  - "Read `main.go`" -> `githubGetFileContent`
- **Best Practices**:
  - Always start with `githubViewRepoStructure` (depth=1 or 2) on new projects to get a mental map.
  - Use `githubSearchCode` before creating new files to avoid duplication.
  - detailed queries in `githubSearchRepositories` help find relevant 3rd party examples if needed.

### 2. Context7 (Documentation)
**Primary Use**: Fetching accurate, up-to-date library documentation.
- **When to use**:
  - "How do I configure `gin-gonic` middleware?"
  - "What are the props for `shadcn/ui` Button?"
- **Workflow**:
  1. `resolve-library-id` (e.g., query="gin")
  2. `get-library-docs` (topic="middleware")

### 3. Supabase (Database & Backend)
**Primary Use**: Managing the project's Supabase backend.
- **When to use**:
  - **SQL**: `execute_sql` (READ), `apply_migration` (WRITE/DDL). *Never* run DDL via raw execute.
  - **Logs**: `get_logs` when debugging API 500 errors or auth failures.
  - **Functions**: `deploy_edge_function` / `get_edge_function` for serverless logic.
- **Security**:
  - Never utilize `execute_sql` with unvalidated user input.
  - Always verify RLS policies using `get_advisors`.

### 4. Playwright (Browser Automation)
**Primary Use**: E2E testing, visual verification, and web scraping.
- **When to use**:
  - "Verify the login page loads" -> `navigate`, `snapshot`.
  - "Take a screenshot of the dashboard" -> `take_screenshot`.
  - "Click the button" -> `click`.
- **Note**: This is the *only* way to "see" or interact with a running web application.

### 5. Shadcn (UI Components)
**Primary Use**: Accelerating React/Tailwind UI development.
- **When to use**:
  - "Add a dialog component" -> `get_add_command_for_items` (returns `npx shadcn@latest add dialog`).
  - "Show me examples of cards" -> `get_item_examples_from_registries`.
- **Workflow**: Search -> View Examples -> Add -> Customize.

### 6. Sequential Thinking (Complex Logic)
**Primary Use**: Deep problem solving, debugging, and architectural planning.
- **When to use**:
  - You are stuck in a loop.
  - The user asks a highly complex architectural question.
  - You need to debug a subtle race condition.
- **Mechanism**: Use `sequentialthinking` to break down the thought process explicitly before taking action.

### 7. MCP Compass (Discovery)
**Primary Use**: Finding *other* tools or servers.
- **When to use**: Rare. Only if the current toolset is insufficient for a specific request (e.g., "I need a tool to query AWS").

## General Rules

1.  **Tool First**: Do not hallucinate file paths or API methods. Use `octocode` or `context7` to verify.
2.  **Safety**: Always check `SafeToAutoRun` constraints on command execution tools.
3.  **Efficiency**: Chain tool calls where possible (e.g., search -> read) to reduce latency, but respect dependency order.
