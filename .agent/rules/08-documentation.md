# Documentation Standards

Guidelines for code comments, API docs, and project documentation.

## Code Comments

### When to Comment
```go
// ✅ Good: Explains WHY (non-obvious reasoning)
// Use exponential backoff to prevent overwhelming the server
// during outages. Max delay is capped at 5 minutes.
delay := min(time.Duration(math.Pow(2, float64(attempt)))*time.Second, 5*time.Minute)

// ✅ Good: Documents complex algorithm
// Uses Kademlia-style XOR distance for peer routing.
// Closer peers (smaller XOR distance) are preferred.
func (r *Router) FindClosestPeers(target ID) []Peer {

// ❌ Bad: States the obvious
// Increment counter by 1
counter++

// ❌ Bad: Outdated comment (worse than no comment)
// Check if user is admin
if user.Role == "moderator" {  // Comment lies!
```

### Comment Style
```go
// Single line comment for brief explanations

/*
Multi-line comment for longer explanations
that need more context or describe complex
behavior spanning multiple paragraphs.
*/
```

---

## Go Documentation (GoDoc)

### Package Comments
```go
// Package service implements the business logic layer for GoConnect.
// It provides network management, user authentication, and peer coordination.
//
// The service layer sits between handlers and repositories, enforcing
// business rules and orchestrating operations across multiple domains.
package service
```

### Function Comments
```go
// CreateNetwork creates a new virtual network with the given configuration.
// It generates a unique network ID, assigns an IP range, and registers
// the network with the coordination server.
//
// The caller must have permission to create networks.
// Returns ErrQuotaExceeded if the user has reached their network limit.
func (s *NetworkService) CreateNetwork(ctx context.Context, req CreateNetworkRequest) (*Network, error) {
```

### Type Comments
```go
// Network represents a virtual LAN that peers can join.
// Each network has a unique ID, a WireGuard configuration,
// and a list of connected peers.
type Network struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    IPRange     netip.Prefix   `json:"ip_range"`
    Peers       []Peer         `json:"peers"`
    CreatedAt   time.Time      `json:"created_at"`
}

// Status returns the current operational status of the network.
func (n *Network) Status() NetworkStatus {
```

---

## API Documentation

### OpenAPI/Swagger
API specs live in `core/openapi/`:

```yaml
# openapi.yaml
openapi: 3.0.3
info:
  title: GoConnect API
  version: 1.0.0
  description: API for managing virtual networks

paths:
  /api/v1/networks:
    post:
      summary: Create a new network
      operationId: createNetwork
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateNetworkRequest'
      responses:
        '201':
          description: Network created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Network'
```

### Generating Docs
```bash
# Generate API docs from OpenAPI spec
cd core
go generate ./...
```

---

## README Guidelines

### Module READMEs
Each module should have a README with:

1. **Overview** - What the module does
2. **Quick Start** - How to build/run
3. **Configuration** - Available options
4. **Architecture** - High-level design (optional)

### Example Structure
```markdown
# GoConnect CLI

Command-line interface and daemon for GoConnect.

## Installation

```bash
go install github.com/orhaniscoding/goconnect/cli/cmd/goconnect@latest
```

## Usage

```bash
# Start daemon
goconnect daemon start

# Join a network
goconnect network join <network-id>

# List connected peers
goconnect peers list
```

## Configuration

Create `~/.goconnect/config.yaml`:

```yaml
server:
  url: https://api.goconnect.io
daemon:
  log_level: info
```
```

---

## CHANGELOG

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Network discovery via mDNS (#78)

### Fixed
- Connection drops on unstable networks (#82)

## [1.1.0] - 2024-01-10

### Added
- File transfer between peers (#45)
- Chat functionality (#52)

### Changed
- Improved connection reliability (#60)

### Deprecated
- Legacy REST API v0 (use v1)

### Removed
- Support for Go 1.21

### Fixed
- Memory leak in WebSocket handler (#55)

### Security
- Updated crypto library for CVE-2024-XXXX
```

---

## Documentation Checklist

When making changes:

- [ ] Code comments explain "why" not "what"
- [ ] Exported functions/types have GoDoc comments
- [ ] README updated if usage changes
- [ ] API changes reflected in OpenAPI spec
- [ ] CHANGELOG updated for user-facing changes
- [ ] Examples updated if API changed
