package main

import (
	"context"
	"fmt"
	"github.com/dist_project/app/log"
	"github.com/dist_project/app/registry"
	"github.com/dist_project/app/service"
	"github.com/dist_project/app/teacherportal"
	stlog "log"
)

func main() {
	err := teacherportal.ImportTemplates()
	if err != nil {
		stlog.Fatal(err)
	}

	host, port := "teacherportal", "5000"

	serviceAddress := fmt.Sprintf("http://%v:%v", host, port)

	var r registry.Registration
	r.ServiceName = registry.TeacherPortal
	r.ServiceURL = serviceAddress
	r.RequiredServices = []registry.ServiceName{
		registry.LogService,
		registry.GradingService,
	}
	r.ServiceUpdateURL = r.ServiceURL + "/services"
	r.HeartbeatURL = r.ServiceURL + "/heartbeat"

	ctx, err := service.Start(context.Background(),
		r.ServiceURL,
		port,
		teacherportal.RegisterHandlers,
		r)

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		log.SetClientLogger(logProvider, r.ServiceName)
	} else {
		stlog.Println("didnt find log provider")
	}

	<-ctx.Done()
	fmt.Println("Shutting down service")

}
