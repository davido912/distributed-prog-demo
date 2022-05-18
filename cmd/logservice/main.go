package main

import (
	"context"
	"fmt"
	"github.com/dist_project/app/log"
	"github.com/dist_project/app/registry"
	"github.com/dist_project/app/service"
	stlog "log"
)

func main() {
	log.Run("./app.log")

	port := "4000"

	var regis registry.Registration
	regis.ServiceName = registry.LogService
	regis.ServiceURL = fmt.Sprintf("http://log:%v", port)
	regis.RequiredServices = []registry.ServiceName{}
	regis.ServiceUpdateURL = regis.ServiceURL + "/services"
	regis.HeartbeatURL = regis.ServiceURL + "/heartbeat"

	ctx, err := service.Start(
		context.Background(),
		regis.ServiceURL,
		port,
		log.RegisterHandlers,
		regis,
	)

	if err != nil {
		stlog.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("Shutting down log service")
}
