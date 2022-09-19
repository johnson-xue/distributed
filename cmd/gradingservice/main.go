package main

import (
	"context"
	"distributed/grades"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	host, port := "localhost", "6000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName:      registry.GradingService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.Logservice},
		ServiceUpdateURL: serviceAddress + "/services",
	}
	ctx, err := service.Start(context.Background(),
		r,
		host,
		port,
		grades.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.Logservice); err == nil {
		fmt.Printf("logging service found at: %s\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}

	<-ctx.Done()
	fmt.Println("Shutting down grading service")
}
