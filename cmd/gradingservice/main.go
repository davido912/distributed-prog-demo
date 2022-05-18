package main

import (
	"context"
	"fmt"
	"github.com/dist_project/app/grades"
	"github.com/dist_project/app/log"
	"github.com/dist_project/app/registry"
	"github.com/dist_project/app/service"
	stlog "log"
)

func main() {
	port := "6000"
	serviceAddress := fmt.Sprintf("http://grading:%v", port)

	var r registry.Registration

	r.ServiceName = registry.GradingService
	r.ServiceURL = serviceAddress
	r.RequiredServices = []registry.ServiceName{
		registry.LogService,
	}
	r.ServiceUpdateURL = r.ServiceURL + "/services"
	r.HeartbeatURL = r.ServiceURL + "/heartbeat"

	ctx, err := service.Start(context.Background(),
		r.ServiceURL,
		port,
		grades.RegisterHandlers,
		r,
	)

	if err != nil {
		stlog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("Logging service found at %v\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)

	}

	<-ctx.Done()
	fmt.Println("Shutting down grading service")
}
