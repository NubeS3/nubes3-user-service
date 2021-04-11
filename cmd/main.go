package main

import (
	"fmt"
	"github.com/Nubes3/common"
	message_queue "github.com/Nubes3/nubes3-user-service/internal/api/message-queue"
	restApi "github.com/Nubes3/nubes3-user-service/internal/api/rest-api"
	"github.com/Nubes3/nubes3-user-service/internal/repo/arangodb"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cleanUps := common.InitCoreComponents(true, false, false, true)
	subCleanUp, err := message_queue.CreateMessageSubcribe()
	if err != nil {
		//TODO log error
		panic(err)
	}

	cleanUps = append(cleanUps, subCleanUp)

	arangodb.InitArangoRepo()

	sigs := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		for _, f := range cleanUps {
			f()
		}
		cleanupDone <- true
	}()

	fmt.Println("User service's running")
	restApi.Serve()
	<-cleanupDone
}
