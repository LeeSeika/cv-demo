package s3

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3bolt"
	"github.com/rs/zerolog/log"
)

var ServerConf ServerConfig

func init() {
	s3DataPath := filepath.Join("tmp", "s3-data")

	// S3 server config
	flag.IntVar(&ServerConf.Port, "s3Port", 4000, "S3 server listening port (default: 4000)")
	flag.BoolVar(&ServerConf.UseHTTPS, "s3Https", false, "Enable HTTPS for S3 server (default: false)")
	flag.StringVar(&ServerConf.CertFile, "s3CertFile", "./cert.pem", "SSL certificate file path (default: ./cert.pem)")
	flag.StringVar(&ServerConf.KeyFile, "s3KeyFile", "./key.pem", "SSL private key file path (default: ./key.pem)")
	flag.StringVar(&ServerConf.DataDir, "s3DataDir", s3DataPath, "Data storage directory (default: ./tmp/s3-data)")
	flag.BoolVar(&ServerConf.AutoCreate, "s3AutoCreate", true, "Automatically create buckets if not exists (default: true)")
	flag.BoolVar(&ServerConf.HostBucket, "s3HostBucket", false, "Use host-style bucket access (default: false)")
	flag.DurationVar(&ServerConf.Timeout, "s3Timeout", 30*time.Second, "Request timeout duration (e.g., 30s, 5m) (default: 30s)")
}

// ServerConfig config for server
type ServerConfig struct {
	Port       int    // port
	UseHTTPS   bool   // use HTTPS
	CertFile   string // SSL file
	KeyFile    string // SSL key
	DataDir    string // data dir
	AutoCreate bool   // auto create
	HostBucket bool
	Timeout    time.Duration // time out
}

type ZeroLog struct{}

func (l ZeroLog) Print(level gofakes3.LogLevel, v ...any) {
	log.Debug().Msgf("%+v", v)
}

func StartS3Server(stop chan int) {
	// init backend
	cfg := ServerConf
	var backend gofakes3.Backend

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatal().Msgf("create %q fail: %v", cfg.DataDir, err)
	}

	var err error
	backend, err = s3bolt.NewFile(
		filepath.Join(cfg.DataDir, "s3.db"),                 // BoltDB path
		s3bolt.WithTimeSource(gofakes3.DefaultTimeSource()), // option
	)
	if err != nil {
		log.Fatal().Msgf("init s3bolt fail: %v", err)
	}

	// init S3 server
	faker := gofakes3.New(backend,
		gofakes3.WithInsecureCORS(),
		gofakes3.WithHostBucket(cfg.HostBucket),
		gofakes3.WithAutoBucket(cfg.AutoCreate),
		gofakes3.WithLogger(ZeroLog{}),
	)

	// created HTTP server
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      faker.Server(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Info().Msgf("mock S3 server on :%d [https: %v]", cfg.Port, cfg.UseHTTPS)

	// start the mock S3 server in a goroutine
	go func() {
		// start server
		if cfg.UseHTTPS {
			if err := server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != http.ErrServerClosed {
				log.Fatal().Msgf("mock S3 server ListenAndServeTLS() error: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatal().Msgf("mock S3 server ListenAndServe() error: %v", err)
			}
		}
	}()

	<-stop

	// wait for 5s to force server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Msg("mock S3 server forced to shutdown")
	}

	log.Info().Msg("mock S3 server is down")
}
