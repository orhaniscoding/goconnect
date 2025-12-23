package service

import (
	"github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/server/internal/daemon"
)

// Program implements service.Interface
type Program struct {
	Daemon *daemon.Daemon
}

// Start starts the service. It must be non-blocking.
func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *Program) run() {
	// Daemon.Run is a blocking call
	if err := p.Daemon.Run(); err != nil {
		// Logger is internal to Daemon
	}
}

// Stop stops the service gracefully.
func (p *Program) Stop(s service.Service) error {
	p.Daemon.Stop()
	return nil
}
