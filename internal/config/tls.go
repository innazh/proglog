package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}

// SetupTLSConfig allows us to get each type of config with one func call, set Server to true to get server's conf, otherwise it for the client
func SetupTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	var err error
	tlsCnf := &tls.Config{}
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsCnf.Certificates = make([]tls.Certificate, 1)
		tlsCnf.Certificates[0], err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	if cfg.CAFile != "" {
		caContents, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM([]byte(caContents))
		if !ok {
			return nil, fmt.Errorf("failed to parse root certificate: %q", cfg.CAFile)
		}
		if cfg.Server {
			tlsCnf.ClientCAs = ca
			tlsCnf.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsCnf.RootCAs = ca
		}
		tlsCnf.ServerName = cfg.ServerAddress
	}
	return tlsCnf, nil
}
