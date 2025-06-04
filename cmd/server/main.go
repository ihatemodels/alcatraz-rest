package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	api "github.com/ihatemodels/alcatraz-rest/internal/api/v1"
	"github.com/ihatemodels/alcatraz-rest/internal/config"
	"github.com/ihatemodels/alcatraz-rest/internal/observability"
)

var version string

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := observability.InitLogger(cfg.Observability)
	logger.Info("starting...", "application", "alcatraz-rest", "version", version)

	// Configure TLS if enabled
	var tlsConfig *tls.Config
	if cfg.Server.TLS.Enabled {
		tlsConfig, err = configureTLS(cfg, logger)
		if err != nil {
			logger.Error("failed to configure TLS", "error", err)
			os.Exit(1)
		}
		logger.Info("TLS configured",
			"cert_file", cfg.Server.TLS.CertFile,
			"require_client_cert", cfg.Server.TLS.RequireClientCert)
	}

	srv := &http.Server{
		Addr:      cfg.GetServerAddress(),
		Handler:   nil,
		TLSConfig: tlsConfig,
	}

	http.HandleFunc("/api/ping", api.PingHandler)

	term := make(chan os.Signal, 1)
	srvClose := make(chan struct{})

	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		var err error
		if cfg.Server.TLS.Enabled {
			logger.Info("starting HTTPS server", "address", cfg.GetServerAddress())
			err = srv.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			logger.Info("starting HTTP server", "address", cfg.GetServerAddress())
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			close(srvClose)
		}
	}()

	for {
		select {
		case <-term:
			logger.Info("Received SIGTERM, exiting gracefully...")
			os.Exit(0)
		case <-srvClose:
			os.Exit(1)
		}
	}
}

// configureTLS sets up TLS configuration including mTLS if required
func configureTLS(cfg *config.Config, logger *slog.Logger) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true,
	}

	// Configure client certificate verification if required
	if cfg.Server.TLS.RequireClientCert {
		logger.Info("configuring mTLS with client certificate verification")

		// Load CA certificate for client verification
		caCert, err := os.ReadFile(cfg.Server.TLS.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read client CA file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse client CA certificate")
		}

		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = caCertPool

		// Add custom verification for debugging
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(verifiedChains) > 0 && len(verifiedChains[0]) > 0 {
				clientCert := verifiedChains[0][0]
				logger.Info("client certificate verified",
					"subject", clientCert.Subject.String(),
					"issuer", clientCert.Issuer.String(),
					"serial", clientCert.SerialNumber.String())
			}
			return nil
		}
	} else {
		tlsConfig.ClientAuth = tls.NoClientCert
	}

	return tlsConfig, nil
}
