package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/instill-ai/component/pkg/base"
	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSLMode is the type for SSL mode
type SSLMode string

const (
	DisableSSLMode    SSLMode = "disable"
	VerifyFullSSLMode SSLMode = "verify-full"
)

// SSLConfig is the interface for SSL configuration
type SSLModeConfig interface {
	GetConfig() (*tls.Config, error)
}

// DisableSSL is the struct for disable SSL
type DisableSSL struct {
	Mode SSLMode `json:"mode"`
}

func (d *DisableSSL) GetConfig() (*tls.Config, error) {
	return nil, nil
}

// VerifyFullSSL is the struct for verify-full SSL. It always requires encryption and verification of the identify of the server.
type VerifyFullSSL struct {
	Mode       SSLMode `json:"mode"`
	CaCert     string  `json:"ca_cert"`
	ClientCert string  `json:"client_cert"`
	ClientKey  string  `json:"client_key"`
}

func (e *VerifyFullSSL) GetConfig() (*tls.Config, error) {
	caCert := []byte(e.CaCert)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// TODO: Add support for password protected client key

	// Load client's certificate and private key
	clientCert, err := tls.X509KeyPair([]byte(e.ClientCert), []byte(e.ClientKey))
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate and key: %v", err)
	}

	// Setup TLS config
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
		// In a production setting, you might want to set MinVersion to tls.VersionTLS12
		MinVersion: tls.VersionTLS12,
		// Setting InsecureSkipVerify to true is not recommended in a production environment
		InsecureSkipVerify: true,
	}
	return tlsConfig, nil
}

func getHost(config *structpb.Struct) string {
	return config.GetFields()["host"].GetStringValue()
}
func getPort(config *structpb.Struct) int {
	return int(config.GetFields()["port"].GetNumberValue())
}
func getPassword(config *structpb.Struct) string {
	val, ok := config.GetFields()["password"]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}
func getUsername(config *structpb.Struct) string {
	val, ok := config.GetFields()["username"]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}

func getSSL(config *structpb.Struct) bool {
	val, ok := config.GetFields()["ssl"]
	if !ok {
		return false
	}
	return val.GetBoolValue()
}

func getSSLMode(config *structpb.Struct) (SSLModeConfig, error) {
	sslMode := config.GetFields()["ssl_mode"].GetStructValue()
	mode := sslMode.GetFields()["mode"].GetStringValue()

	var sslModeConfig SSLModeConfig
	switch mode {
	case string(DisableSSLMode):
		sslModeConfig = &DisableSSL{}
	case string(VerifyFullSSLMode):
		sslModeConfig = &VerifyFullSSL{}
	default:
		return nil, fmt.Errorf("invalid SSL mode: %s", mode)
	}

	err := base.ConvertFromStructpb(sslMode, sslModeConfig)
	if err != nil {
		return nil, err
	}
	return sslModeConfig, nil
}

// NewClient creates a new redis client
func NewClient(config *structpb.Struct) (*goredis.Client, error) {
	op := &goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", getHost(config), getPort(config)),
		Password: getPassword(config),
		DB:       0,
	}
	if getUsername(config) != "" {
		op.Username = getUsername(config)
	}

	if getSSL(config) {
		sslConfig, err := getSSLMode(config)
		if err != nil {
			return nil, err
		}
		if sslConfig != nil {
			tlsConfig, err := sslConfig.GetConfig()
			if err != nil {
				return nil, err
			}
			op.TLSConfig = tlsConfig
		}
	}

	// TODO - add SSH support

	return goredis.NewClient(op), nil
}
