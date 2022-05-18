package main

import (
	"context"
	"fmt"
	"github.com/dist_project/app/registry"
	"log"
	"net/http"
)

func main() {
	registry.SetupRegistryService()
	http.Handle("/services", registry.RegistryService{})

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	var srv http.Server
	srv.Addr = registry.ServerPort

	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()
	fmt.Printf("Registry service started on %v.\n", srv.Addr)
	//go func() {
	//	fmt.Printf("Registry service started on %v. Press any key to stop.\n", srv.Addr)
	//	//var s string
	//	//fmt.Scanln(&s)
	//	//srv.Shutdown(ctx)
	//	cancel()
	//}()

	<-ctx.Done()

	fmt.Println("shutting down registry service")

}
