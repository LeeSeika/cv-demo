package config

type S3Env struct {
	IsAuth          bool   `env:"IS_AUTH" envDefault:"true"`        // Whether authentication is needed (authentication is need for production cloud S3 provider)
	Bucket          string `env:"BUCKET" envDefault:""`             // S3 Bucket
	AccessKeyID     string `env:"ACCESS_KEY_ID" envDefault:""`      // S3 Access key
	SecretAccessKey string `env:"SECRET_ACCESS_KEY" envDefault:""`  // S3 Secret access key
	Region          string `env:"REGION" envDefault:""`             // S3 Region
	Endpoint        string `env:"ENDPOINT" envDefault:""`           // S3 Endpoint
	URLScheme       string `env:"URL_SCHEME" envDefault:""`         // URL scheme (`https` or `http`)
	UsePathStyle    bool   `env:"USE_PATH_STYLE" envDefault:"true"` // Allow you to enable the client to use path-style addressing, e.g. https://s3.amazonaws.com/BUCKET/KEY
}

type S3Config struct {
	IsAuth          bool   `json:"is_auth"`           // Whether authentication is needed (authentication is need for production cloud S3 provider)
	Bucket          string `json:"bucket"`            // S3 Bucket
	AccessKeyID     string `json:"access_key_id"`     // S3 Access key
	SecretAccessKey string `json:"secret_access_key"` // S3 Secret access key
	Region          string `json:"region"`            // S3 Region
	Endpoint        string `json:"endpoint"`          // S3 Endpoint
	URLScheme       string `json:"url_scheme"`        // URL scheme (`https` or `http`)
	UsePathStyle    bool   `json:"use_path_style"`    // Allow you to enable the client to use path-style addressing, e.g. https://s3.amazonaws.com/BUCKET/KEY
}
