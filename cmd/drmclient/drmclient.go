package main

import (
	"log"

	"github.com/titosilva/drmchain-pos/cmd/drmclient/internal/cmdserver"
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/di/defaultdi"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/networkdi"
)

func main() {
	log.Println("drmchain-pos client")

	log.Println("Configuring DI context")
	diCtx := defaultdi.ConfigureDefaultDI()
	log.Println("DI context configured")

	log.Println("Reading self identity")
	idProvider := identityprovider.GetFromDI(diCtx)

	self, err := idProvider.GetIdentity()
	if err != nil {
		log.Printf("Error getting identity: %v\n", err)
		return
	}
	log.Printf("Self identity tag: %v", self.GetTag())

	log.Println("Starting command server")
	cmdServer := cmdserver.NewCommandsServer()
	cmdServer.Start()
	log.Println("Command server started")

	log.Println("Starting network services")
	diCtx = networkdi.AddNetworkServices(diCtx)
	nw := di.GetService[network.Network](diCtx)
	if err = nw.Open(); err != nil {
		log.Println("failed to initialize network services", err)
		return
	}
	defer nw.Close()

	currentConns := nw.GetConnections().Current()
	for conn := range currentConns.All() {
		log.Println(conn.GetPeer().Addr)
	}

	log.Println("Exiting drmchain-pos CLI tool")
}
