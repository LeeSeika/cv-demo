package config

import "fmt"

type StorageEnv struct {
	OSSType string `env:"TYPE" envDefault:""` // Storage type
	S3      S3Env  `envPrefix:"S3."`          // S3 configuration
	GCS     GCSEnv `envPrefix:"GCS."`         // Google Cloud Storage configuration
}

func (env *StorageEnv) Check() (StorageConfig, error) {
	s3 := S3Config{
		IsAuth:          env.S3.IsAuth,
		AccessKeyID:     env.S3.AccessKeyID,
		SecretAccessKey: env.S3.SecretAccessKey,
		Region:          env.S3.Region,
		Endpoint:        env.S3.Endpoint,
		URLScheme:       env.S3.URLScheme,
		UsePathStyle:    env.S3.UsePathStyle,
		Bucket:          env.S3.Bucket,
	}

	gcsCredentials := GCSCredentialsConfig{
		Type:                    env.GCS.Credentials.Type,
		ProjectID:               env.GCS.Credentials.ProjectID,
		PrivateKeyID:            env.GCS.Credentials.PrivateKeyID,
		PrivateKey:              env.GCS.Credentials.PrivateKey,
		ClientEmail:             env.GCS.Credentials.ClientEmail,
		ClientID:                env.GCS.Credentials.ClientID,
		AuthURI:                 env.GCS.Credentials.AuthURI,
		TokenURI:                env.GCS.Credentials.TokenURI,
		AuthProviderX509CertURL: env.GCS.Credentials.AuthProviderX509CertURL,
		ClientX509CertURL:       env.GCS.Credentials.ClientX509CertURL,
	}

	g := GCSConfig{
		IsAuth:      env.GCS.IsAuth,
		URLScheme:   env.GCS.URLScheme,
		Host:        env.GCS.Host,
		Port:        env.GCS.Port,
		Bucket:      env.GCS.Bucket,
		Credentials: gcsCredentials,
	}

	storageConfig := StorageConfig{
		OSSType: env.OSSType,
		S3:      s3,
		GCS:     g,
	}

	if err := storageConfig.check(); err != nil {
		return StorageConfig{}, err
	}

	return storageConfig, nil
}

type StorageConfig struct {
	OSSType string    `json:"type"` // Cloud Storage type (either `s3` or `gcs`)
	S3      S3Config  `json:"s3"`   // S3 configuration
	GCS     GCSConfig `json:"gcs"`  // Google Cloud Storage configuration
}

func (cfg *StorageConfig) check() error {
	if cfg == nil {
		return fmt.Errorf("nil StorageConfig")
	}

	switch cfg.OSSType {
	case "":
		{
			return fmt.Errorf("missing storage TYPE")
		}
	case "s3":
		{
			// we have to define bucket
			if cfg.S3.Bucket == "" {
				return fmt.Errorf("S3.BUCKET is required")
			}

			if cfg.S3.AccessKeyID == "" {
				return fmt.Errorf("S3.ACCESS_KEY_ID is required")
			}

			if cfg.S3.SecretAccessKey == "" {
				return fmt.Errorf("S3.SECRET_ACCESS_KEY is required")
			}

			if cfg.S3.Region == "" {
				return fmt.Errorf("S3.REGION is required")
			}

			if cfg.S3.Endpoint == "" {
				return fmt.Errorf("S3.ENDPOINT is required")
			}

			if cfg.S3.URLScheme == "" {
				return fmt.Errorf("S3.URL_SCHEME is required")
			}

			if cfg.S3.URLScheme != "https" && cfg.S3.URLScheme != "http" {
				return fmt.Errorf("S3.URL_SCHEME is incorrect (either `https` or `http`)")
			}
		}
	case "gcs":
		{
			// we have to define bucket
			if cfg.GCS.Bucket == "" {
				return fmt.Errorf("GCS.BUCKET is required")
			}

			if cfg.GCS.IsAuth {
				if cfg.GCS.Credentials.Type == "" {
					return fmt.Errorf("GCS.CREDENTIALS.TYPE is required")
				}

				if cfg.GCS.Credentials.ProjectID == "" {
					return fmt.Errorf("GCS.CREDENTIALS.PROJECT_ID is required")
				}

				if cfg.GCS.Credentials.PrivateKeyID == "" {
					return fmt.Errorf("GCS.CREDENTIALS.PRIVATE_KEY_ID is required")
				}

				if cfg.GCS.Credentials.PrivateKey == "" {
					return fmt.Errorf("GCS.CREDENTIALS.PRIVATE_kEY is required")
				}

				if cfg.GCS.Credentials.ClientEmail == "" {
					return fmt.Errorf("GCS.CREDENTIALS.CLIENT_EMAIL is required")
				}

				if cfg.GCS.Credentials.ClientID == "" {
					return fmt.Errorf("GCS.CREDENTIALS.CLIENT_ID is required")
				}

				if cfg.GCS.Credentials.AuthURI == "" {
					return fmt.Errorf("GCS.CREDENTIALS.AUTH_URI is required")
				}

				if cfg.GCS.Credentials.TokenURI == "" {
					return fmt.Errorf("GCS.CREDENTIALS.TOKEN_URI is required")
				}

				if cfg.GCS.Credentials.AuthProviderX509CertURL == "" {
					return fmt.Errorf("GCS.CREDENTIALS.AUTH_PROVIDER_X509_CERT_URL is required")
				}

				if cfg.GCS.Credentials.ClientX509CertURL == "" {
					return fmt.Errorf("GCS.CREDENTIALS.CLIENT_X509_CERT_URL is required")
				}
			} else {
				if cfg.GCS.URLScheme == "" {
					return fmt.Errorf("GCS.URL_SCHEME is required")
				}

				if cfg.GCS.URLScheme != "https" && cfg.GCS.URLScheme != "http" {
					return fmt.Errorf("GCS.URL_SCHEME is incorrect (either `https` ot `http`)")
				}

				if cfg.GCS.Host == "" {
					return fmt.Errorf("GCS.HOST cannot be empty")
				}

				if cfg.GCS.Port == "" {
					return fmt.Errorf("GCS.PORT cannot be empty")
				}
			}
		}
	default:
		return fmt.Errorf("unknown storage TYPE: should be 's3' or 'gcs'")
	}

	return nil
}
