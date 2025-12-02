package service

import (
	"log"

	"github.com/kardianos/service"
)

// Program implements the service.Interface
type Program struct {
	exit chan struct{}
}

// Start is called when the service is started
func (p *Program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

// run is the main loop of the service
func (p *Program) run() {
	log.Println("GoConnect Daemon started")
	// TODO: Initialize gRPC server, SQLite store, etc.
	
	// Keep the service running until Stop is called
	<-p.exit
	log.Println("GoConnect Daemon stopping...")
}

// Stop is called when the service is stopped
func (p *Program) Stop(s service.Service) error {
	// Stop should not block. Signal the exit channel.
	close(p.exit)
	return nil
}
