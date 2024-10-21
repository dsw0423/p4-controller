package main

import (
	"log"
	"os"

	"github.com/antoninbas/p4runtime-go-client/pkg/client"
	"github.com/antoninbas/p4runtime-go-client/pkg/signals"
	"github.com/gin-gonic/gin"
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultDeviceId               = 1
	defalutP4RuntimeServerAddress = "127.0.0.1:9559"
	defaultWebServerAddress       = ":8080"
)

var (
	/* are we the primary controller? */
	isPrimary bool
	/* temp directory saving files. */
	tmpDir string
	/* P4Runtime controller. */
	p4rt_ctl *client.Client
)

func main() {
	initialize()

	/* start p4rt_ctl */
	conn, err := grpc.NewClient(defalutP4RuntimeServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	stub := p4_v1.NewP4RuntimeClient(conn)
	electionId := &p4_v1.Uint128{High: 0, Low: 1}
	p4rt_ctl = client.NewClient(stub, defaultDeviceId, electionId)
	arbitrationCh := make(chan bool)
	stopCh := signals.RegisterSignalHandlers()

	go p4rt_ctl.Run(stopCh, arbitrationCh, nil)

	/* monitoring arbitration result */
	go monitoringArbitration(arbitrationCh)

	/* start web router */
	router := gin.Default()

	// setting pipline config
	router.POST("/pipeconf", setPipeconf)
	// insert a table entry using exact matching
	router.POST("/tableEntryExact", insertTableEntryExact)

	go router.Run(defaultWebServerAddress)
	<-stopCh
	log.Println("stopping...")
}

func monitoringArbitration(arbitrationCh chan bool) {
	for primary := range arbitrationCh {
		isPrimary = primary
		if isPrimary {
			log.Println("we are the primary controller.")
		} else {
			log.Println("we are NOT the primary controller.")
		}
	}
}

func initialize() {
	tmpDir, _ = os.Getwd()
	tmpDir = tmpDir + "/tmp/"
	log.Printf("tmpDir: %s\n", tmpDir)
}
