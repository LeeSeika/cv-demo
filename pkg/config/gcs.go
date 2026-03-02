package config

type GCSEnv struct {
	IsAuth      bool              `env:"IS_AUTH" envDefault:"true"` // Whether authentication is needed (use Credentials JSON instead of specifying endpoint API)
	URLScheme   string            `env:"URL_SCHEME" envDefault:""`  // URL scheme (`https` or `http`) (only valid if IS_AUTH is false)
	Host        string            `env:"HOST" envDefault:""`        // Google Cloud Storage API host (only valid if IS_AUTH is false)
	Port        string            `env:"PORT" envDefault:""`        // Google Cloud Storage API port (only valid if IS_AUTH is false)
	Bucket      string            `env:"BUCKET" envDefault:""`      // Google Cloud Storage Bucket
	Credentials GCSCredentialsEnv `envPrefix:"CREDENTIALS."`        // Google Cloud Storage Credentials
}

type GCSCredentialsEnv struct {
	Type                    string `env:"TYPE" envDefault:""`                        // The type of authentication, this should always be `service_account`.
	ProjectID               string `env:"PROJECT_ID" envDefault:""`                  // The name of the current project.
	PrivateKeyID            string `env:"PRIVATE_KEY_ID" envDefault:""`              // A unqiue identifier for the private key.
	PrivateKey              string `env:"PRIVATE_KEY" envDefault:""`                 // The private key in RSA format.
	ClientEmail             string `env:"CLIENT_EMAIL" envDefault:""`                // The email address associated with the service account.
	ClientID                string `env:"CLIENT_ID" envDefault:""`                   // The unique identifier for this client.
	AuthURI                 string `env:"AUTH_URI" envDefault:""`                    // The endpoint where authentication happens.
	TokenURI                string `env:"TOKEN_URI" envDefault:""`                   // The endpoint where OAuth2 tokens are issued.
	AuthProviderX509CertURL string `env:"AUTH_PROVIDER_X509_CERT_URL" envDefault:""` // The url of the cert provider.
	ClientX509CertURL       string `env:"CLIENT_X509_CERT_URL" envDefault:""`        // The url of a static file containing metadata for this certificate.
}

type GCSConfig struct {
	IsAuth      bool                 `json:"is_auth"`     // Whether authentication is needed (use Credentials JSON instead of specifying endpoint API)
	URLScheme   string               `json:"url_scheme"`  // URL scheme (`https` or `http`) (only valid if IS_AUTH is false)
	Host        string               `json:"host"`        // Google Cloud Storage API host (only valid if IS_AUTH is false)
	Port        string               `json:"port"`        // Google Cloud Storage API port (only valid if IS_AUTH is false)
	Bucket      string               `json:"bucket"`      // Google Cloud Storage Bucket
	Credentials GCSCredentialsConfig `json:"credentials"` // Google Cloud Storage Credentials
}

type GCSCredentialsConfig struct {
	Type                    string `json:"type"`                        // The type of authentication, this should always be `service_account`.
	ProjectID               string `json:"project_id"`                  // The name of the current project.
	PrivateKeyID            string `json:"private_key_id"`              // A unqiue identifier for the private key.
	PrivateKey              string `json:"private_key"`                 // The private key in RSA format.
	ClientEmail             string `json:"client_email"`                // The email address associated with the service account.
	ClientID                string `json:"client_id"`                   // The unique identifier for this client.
	AuthURI                 string `json:"auth_uri"`                    // The endpoint where authentication happens.
	TokenURI                string `json:"token_uri"`                   // The endpoint where OAuth2 tokens are issued.
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"` // The url of the cert provider.
	ClientX509CertURL       string `json:"client_x509_cert_url"`        // The url of a static file containing metadata for this certificate.
}
