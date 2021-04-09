package main

import (
	"fmt"
	"github.com/Nubes3/common"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cleanUps := common.InitCoreComponents(true, false, false, false)

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
	<-cleanupDone
}
