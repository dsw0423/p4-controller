package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dsw0423/p4-controller/internal/handler"
	log "github.com/sirupsen/logrus"

	"github.com/antoninbas/p4runtime-go-client/pkg/client"
	"github.com/antoninbas/p4runtime-go-client/pkg/signals"
	"github.com/gin-gonic/gin"
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"github.com/redis/go-redis/v9"
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
	/* Redis client. */
	redisClient *redis.Client
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
	electionId := &p4_v1.Uint128{High: 0, Low: 100}
	p4rt_ctl = client.NewClient(stub, defaultDeviceId, electionId)
	arbitrationCh := make(chan bool)
	messageCh := make(chan *p4_v1.StreamMessageResponse, 100)
	stopCh := signals.RegisterSignalHandlers()

	go p4rt_ctl.Run(stopCh, arbitrationCh, messageCh)

	/* handle StreamChannel messages except arbitration result. */
	go func() {
		ctx := context.Background()
		handleStreamMessages(ctx, p4rt_ctl, messageCh)
	}()

	/* monitoring arbitration result */
	go monitoringArbitration(arbitrationCh)

	/* start web router */
	router := gin.Default()

	// login
	router.POST("/login", handler.Login)

	// refresh tokens
	router.POST("/refreshToken", handler.RefreshToken)

	authGroup := router.Group("/auth", handler.AuthCheck)
	{
		// setting pipline config
		authGroup.POST("/pipeconf", setPipeconfHandler)
		// insert a table entry using exact matching
		authGroup.POST("/tableEntryExact", insertTableEntryExactHandler)
		// send a PacketOut stream message
		authGroup.POST("/packetout", sendPacketOutHandler)
		// get table entries by name
		authGroup.GET("/tableEntries", getTableEntriesByNameHandler)
	}

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

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

}

func handleStreamMessages(ctx context.Context, p4RtC *client.Client, messageCh <-chan *p4_v1.StreamMessageResponse) {
	for message := range messageCh {
		switch message.Update.(type) {
		case *p4_v1.StreamMessageResponse_Packet:
			log.Debugf("Received PacketIn")
			packet := message.Update.(*p4_v1.StreamMessageResponse_Packet)
			for _, metadata := range packet.Packet.Metadata {
				fmt.Printf("metadata ID: %v, value: %v\n", metadata.MetadataId, metadata.Value)
			}
		case *p4_v1.StreamMessageResponse_Digest:
			log.Debugf("Received DigestList")
			/* if err := learnMacs(ctx, p4RtC, m.Digest); err != nil {
				log.Errorf("Error when learning MACs: %v", err)
			} */
		case *p4_v1.StreamMessageResponse_IdleTimeoutNotification:
			log.Debugf("Received IdleTimeoutNotification")
			// forgetEntries(ctx, p4RtC, m.IdleTimeoutNotification)
		case *p4_v1.StreamMessageResponse_Error:
			log.Errorf("Received StreamError")
		default:
			log.Errorf("Received unknown stream message")
		}
	}
}
