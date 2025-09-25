GoConnect Configuration Flags

This document lists environment flags and their effects. Unless stated otherwise, flags are read at process start.

Server (goconnect-server)

- SERVER_RL_CAPACITY
	- Type: integer
	- Default: 5
	- Description: Number of requests allowed per window for a given key (user_id if authenticated, else client IP) by the built-in rate limiter applied to mutation endpoints.

- SERVER_RL_WINDOW_MS
	- Type: integer (milliseconds)
	- Default: 1000 (1s)
	- Description: Window size for the built-in rate limiter.

Notes

- The rate limiter is applied to mutation endpoints (create network, join/approve/deny/kick/ban, update/delete) to protect the API. Reads are not rate-limited by default.
- On limit exceed, the server returns HTTP 429 with a standard error body and sets the Retry-After header (seconds). CORS exposes Retry-After for browser clients.

Client-Daemon & Web UI

- TBD. This file will evolve as more flags are introduced.
features.json: beta_webchat, relay_enabled...
