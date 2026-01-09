# Feature: Network Join Flow

## Raw Description
Users can join networks via invite codes or `gc://` protocol deep links, with automatic WireGuard tunnel configuration.

## Source
Product Roadmap - Phase 1: Core Networking (MVP), Item #2

## Scope Confirmation
This feature covers:
- Joining networks via 8-character invite codes
- Deep link support (`gc://join?code=XYZ`)
- Automatic WireGuard tunnel setup after joining
- Join request flow for approval-based networks
- UI feedback for join success/failure states

## Related Features
- **Depends on**: Network Creation & Management (completed)
- **Related to**: Peer Discovery & Connection (next)
