package gcs

import (
	"flag"
	"path/filepath"
	"time"

	"github.com/leeseika/cv-demo/integration/gcs/mock/fakestorage"
	"github.com/rs/zerolog/log"
)

func init() {
	gcsDataPath := filepath.Join("tmp", "gcs-data")

	flag.StringVar(&GCSCfg.Host, "gcsHost", "0.0.0.0", "Server host (default: 0.0.0.0)")
	flag.IntVar(&GCSCfg.Port, "gcsPort", 4001, "Server port (default: 4001)")
	flag.StringVar(&GCSCfg.Scheme, "gcsScheme", "http", "Protocol scheme (http | https)")
	flag.StringVar(&GCSCfg.StorageRoot, "gcsData", gcsDataPath, "Persistence directory")
	flag.StringVar(&GCSCfg.ExternalURL, "gcsExternalURL", "http://localhost:4001", "External access URL")
	flag.StringVar(&GCSCfg.PublicHost, "gcsPublicHost", "127.0.0.1:4001", "Public hostname")
	flag.StringVar(&GCSCfg.CORSHeaders, "gcsCORS", "*", "Allowed CORS headers (comma separated)")
	flag.StringVar(&GCSCfg.LogLevel, "gcsLog", "info", "Log level (debug | info | warn |error)")
	flag.BoolVar(&GCSCfg.EnableVersioning, "gcsVersioning", false, "Enable object versioning")
	flag.DurationVar(&GCSCfg.Timeout, "gcsTimeout", 30*time.Second, "Request timeout")
	flag.BoolVar(&GCSCfg.MemoryBackend, "gcsMemory", false, "Use memory backend")
	flag.StringVar(&GCSCfg.CertFile, "gcsCert", "", "SSL certificate path")
	flag.StringVar(&GCSCfg.KeyFile, "gcsKey", "", "SSL private key path")
	flag.StringVar(&GCSCfg.Bucket, "gcsBucket", "cv-demo", "Bucket name (default: cv-demo)")
}

var GCSCfg GCSConfig

// GCSConfig server config
type GCSConfig struct {
	Host               string
	Port               int
	Scheme             string
	StorageRoot        string
	ExternalURL        string
	PublicHost         string
	CORSHeaders        string
	LogLevel           string
	EnableVersioning   bool
	Timeout            time.Duration
	MemoryBackend      bool
	CertFile           string
	KeyFile            string
	Bucket             string
	DefaultFile        string
	DefaultFileContent string
	ContentType        string
}

func StartGCSServer(stop chan int, publicHost string) {
	// create `test1.txt` in the default bucket
	defaultFile := "test1.txt"
	defaultContentType := "text/plain"
	defaultFileContent := "hello GCS"

	opts := fakestorage.Options{
		Host:               GCSCfg.Host,
		Port:               uint16(GCSCfg.Port),
		Scheme:             GCSCfg.Scheme,
		StorageRoot:        GCSCfg.StorageRoot,
		ExternalURL:        GCSCfg.ExternalURL,
		PublicHost:         publicHost,
		AllowedCORSHeaders: []string{GCSCfg.CORSHeaders},
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName:  GCSCfg.Bucket,
					Name:        defaultFile,
					ContentType: defaultContentType,
					Metadata: map[string]string{
						"source": "gcs",
					},
				},
				Content: []byte(defaultFileContent),
			},
		},
	}

	server, err := fakestorage.NewServerWithOptions(opts)
	if err != nil {
		log.Fatal().Msgf("failed to start mock GCS server: %v", err)
	}
	defer server.Stop()

	log.Info().Msgf("mock GCS server running at: %s", server.URL())

	// wait for stop signal
	<-stop

	log.Info().Msg("mock GCS server is down")
}
