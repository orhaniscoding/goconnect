package main

import (
	"flag"
	"log"

	kardianos "github.com/kardianos/service"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

func main() {
	svcConfig := &kardianos.Config{
		Name:        "GoConnect",
		DisplayName: "GoConnect Daemon",
		Description: "Core daemon for the GoConnect network overlay.",
	}

	prg := &service.Program{}
	s, err := kardianos.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Define command-line flags
	serviceFlag := flag.String("service", "", "Control the system service: install, uninstall, start, stop")
	flag.Parse()

	if len(*serviceFlag) != 0 {
		err := kardianos.Control(s, *serviceFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", kardianos.ControlAction)
			log.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
