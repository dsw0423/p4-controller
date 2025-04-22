package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
	/* temp directory saving files. */
	tmpDir string
	/* Redis client. */
	redisClient *redis.Client

	hostsInfo map[string]*HostInfo

	stopCh <-chan struct{}
)

func main() {
	initialize()

	router := gin.Default()
	router.Use(handler.Cors)
	router.POST("/login", handler.Login)
	router.POST("/refreshToken", handler.RefreshToken)
	router.GET("/portsBitRate", portsBitRateHandler)
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

		// get file list
		authGroup.GET("/filesList", filesListHandler)

		// delete file by hash
		authGroup.DELETE("/file/:hash", fileDeleteHandler)

		// get file by hash
		authGroup.GET("/file/:hash", fileDownloadHandler)

		authGroup.DELETE("/tableEntry", deleteTableEntryHandler)

		authGroup.GET("/p4info", getP4InfoHandler)

		authGroup.GET("/portsInfo", getPortsInfoAndStatusHandler)
	}

	go router.Run(defaultWebServerAddress)
	<-stopCh
	log.Println("stopping...")
}

func initialize() {
	if os.Geteuid() != 0 {
		log.Errorln("root permission is required.")
		os.Exit(1)
	}

	tmpDir, _ = os.Getwd()
	tmpDir = tmpDir + "/tmp/"
	log.Printf("tmpDir: %s\n", tmpDir)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	/* TODO: 写到配置文件里 */
	hostsInfo = map[string]*HostInfo{
		"0": {
			IP:            "127.0.0.1",
			P4RuntimePort: "9559",
			HTTPPort:      "8089",
		},
	}

	stopCh = signals.RegisterSignalHandlers()

	for _, host := range hostsInfo {
		/* start p4rt_ctl for each host */
		conn, err := grpc.NewClient(host.IP+":"+host.P4RuntimePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		arbitrationCh := make(chan bool)
		messageCh := make(chan *p4_v1.StreamMessageResponse, 100)
		stub := p4_v1.NewP4RuntimeClient(conn)
		electionId := &p4_v1.Uint128{High: 0, Low: 100}
		host.P4RTClient = client.NewClientForRole(stub, defaultDeviceId, electionId, &p4_v1.Role{Name: "main"})

		go func() {
			for {
				if err := host.P4RTClient.Run(stopCh, arbitrationCh, messageCh); err == nil {
					break
				} else {
					log.Println(err.Error())
				}
				log.Println("Trying to reconnect to P4Runtime server in 100ms...")
				time.Sleep(100 * time.Millisecond)
			}
		}()

		/* handle StreamChannel messages except arbitration result. */
		go func() {
			ctx := context.Background()
			handleStreamMessages(ctx, host.P4RTClient, messageCh)
		}()

		/* monitoring arbitration result */
		go monitoringArbitration(host, arbitrationCh)
	}
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

func monitoringArbitration(host *HostInfo, arbitrationCh chan bool) {
	for primary := range arbitrationCh {
		host.IsPrimary = primary
		if host.IsPrimary {
			log.Printf("we are the primary controller for %s", host.IP)
		} else {
			log.Printf("we are NOT the primary controller for %s", host.IP)
		}
	}
}
