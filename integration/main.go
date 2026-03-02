/*
Copyright: 2026, Deep Codify Limited

intercart
*/

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leeseika/cv-demo/integration/gcs"
	"github.com/leeseika/cv-demo/integration/pubsub"
	"github.com/leeseika/cv-demo/integration/s3"
	"github.com/rs/zerolog/log"
)

var (
	enableS3      = flag.Bool("s3", true, "whether to enable mock S3 service")
	enableGCS     = flag.Bool("gcs", true, "whether to enable mock Google Cloud Storage service")
	gcsPublicHost = flag.String("gcsPubHost", "127.0.0.1", "gcs public host (default: '127.0.0.1')")
	enablePubsub  = flag.Bool("pubsub", false, "whether to enable pubsub server")
	pubsubPort    = flag.Int("pubsubPort", 8085, "Pub/sub Server TCP port number to listen on (default:8085)")
)

func main() {
	log.Info().Msg("integration server")

	// start pub/sub
	if *enablePubsub {
		pubsubPortUsed := 8085
		if pubsubPort != nil {
			pubsubPortUsed = *pubsubPort
		}
		err := pubsub.StartTestServer(pubsubPortUsed)
		if err != nil {
			log.Fatal().Msgf("failed to start pub/sub server: %+v", err)
		}
	}

	// whether to start mock s3 service.
	stopS3 := make(chan int, 1)
	if *enableS3 {
		go s3.StartS3Server(stopS3)
	}

	// whether to start mock Google Cloud Storage service.
	stopGCS := make(chan int, 1)
	if *enableGCS {
		gcsPublicHostUsed := "127.0.0.1"
		if gcsPublicHost != nil {
			gcsPublicHostUsed = *gcsPublicHost
		}
		go gcs.StartGCSServer(stopGCS, gcsPublicHostUsed)
	}

	// shut down on signal
	stopMainThread := make(chan os.Signal, 1)
	signal.Notify(stopMainThread, syscall.SIGINT, syscall.SIGTERM)

	// blocks
	<-stopMainThread

	if *enableS3 {
		log.Info().Msgf("shutting down mock S3 Server")
		stopS3 <- 1
	}

	time.Sleep(1 * time.Second)

	if *enableGCS {
		log.Info().Msgf("shutting down mock GCS Server")
		stopGCS <- 1
	}

	time.Sleep(1 * time.Second)

	if *enablePubsub {
		pubsub.CloseTestServer()
	}

	log.Info().Msgf("integration is shut down")

}
