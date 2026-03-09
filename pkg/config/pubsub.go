package config

// PubSubConf Event Stream Configuration.
type PubSubConf struct {
	Provider         string           `json:"provider"`
	GooglePubsubConf GooglePubsubConf `json:"google_pubsub_conf"`
}

type GooglePubsubConf struct {
	UseMock           bool              `json:"use_mock"`
	MockPubsub        MockPubsubConf    `json:"mock_pubsub"`
	GPSServiceAccount GPSServiceAccount `json:"gps_service_account"`
	PublisherConfig   PublisherConfig   `json:"publisher_config"`
	SubscriberConfig  SubscriberConfig  `json:"subscriber_config"`
}

type GPSServiceAccount struct {
	Type                    string `json:"type"`                        // Google Cloud Storage: The type of authentication, this should always be `service_account`
	ProjectID               string `json:"project_id"`                  // Google Cloud Storage: The name of the current project
	PrivateKeyID            string `json:"private_key_id"`              // Google Cloud Storage: A unqiue identifier for the private key
	PrivateKey              string `json:"private_key"`                 // Google Cloud Storage: The private key in RSA format
	ClientEmail             string `json:"client_email"`                // Google Cloud Storage: The email address associated with the service account
	ClientID                string `json:"client_id"`                   // Google Cloud Storage: The unique identifier for this client
	AuthURI                 string `json:"auth_uri"`                    // Google Cloud Storage: The endpoint where authentication happens
	TokenURI                string `json:"token_uri"`                   // Google Cloud Storage: The endpoint where OAuth2 tokens are issued
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"` // Google Cloud Storage: The url of the cert provider
	ClientX509CertURL       string `json:"client_x509_cert_url"`        // Google Cloud Storage: The url of a static file containing metadata for this certificate
}

type MockPubsubConf struct {
	Endpoint  string `json:"endpoint"`
	ProjectID string `json:"project_id"`
}

type PublisherConfig struct {
	Topics []string `json:"topics"`
}

type SubscriberConfig struct {
	SubscriberID        string `json:"subscriber_id"`
	MaxDeliveryAttempts int    `json:"max_delivery_attempts"`
	MinimumBackoff      int    `json:"minimum_backoff"`
	MaximumBackoff      int    `json:"maximum_backoff"`
}
