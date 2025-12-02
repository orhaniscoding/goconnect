package store

import "context"

// Node represents the local node's identity and configuration
type Node struct {
	ID         string
	PublicKey  string
	PrivateKey string
	Name       string
}

// Peer represents a remote node in the network
type Peer struct {
	ID         string
	PublicKey  string
	Endpoint   string
	AllowedIPs string
}

// Store defines the interface for persistence operations
type Store interface {
	// Node operations
	GetNode(ctx context.Context) (*Node, error)
	SaveNode(ctx context.Context, node *Node) error

	// Peer operations
	ListPeers(ctx context.Context) ([]*Peer, error)
	SavePeer(ctx context.Context, peer *Peer) error
	DeletePeer(ctx context.Context, publicKey string) error

	// Lifecycle
	Close() error
}
