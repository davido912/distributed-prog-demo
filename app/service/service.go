package service

import (
	"context"
	"fmt"
	"github.com/dist_project/app/registry"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Start(ctx context.Context, host, port string, registerHandlersFunc func(), reg registry.Registration) (context.Context, error){

	registerHandlersFunc()
	ctx = startService(ctx, reg.ServiceName, host, port)

	err := registry.RegisterService(reg)

	if err != nil {
		return ctx, err
	}

	return ctx, nil

}

func startService(ctx context.Context, serviceName registry.ServiceName, host, port string)  context.Context {

	ctx, cancel := context.WithCancel(ctx)

	var srv http.Server
	srv.Addr = ":" + port

	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()
	fmt.Printf("%v started. \n", serviceName)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGTERM)



	go func() {
		sig := <- sigs
		fmt.Println("Service " + host +  "is shutting down: ", sig)

		//var s string
		//fmt.Scanln(&s)
		err := registry.ShutdownService(host)
		if err != nil {
			log.Println(err)
		}
		srv.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
