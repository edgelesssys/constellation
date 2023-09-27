/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package main parses command line flags and starts the s3proxy server.
*/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/s3proxy/internal/router"
	"go.uber.org/zap"
)

const (
	// defaultPort is the default port to listen on.
	defaultPort = 4433
	// defaultIP is the default IP to listen on.
	defaultIP = "0.0.0.0"
	// defaultRegion is the default AWS region to use.
	defaultRegion = "eu-west-1"
	// defaultCertLocation is the default location of the TLS certificate.
	defaultCertLocation = "/etc/s3proxy/certs"
	// defaultLogLevel is the default log level.
	defaultLogLevel = 0
)

func main() {
	flags, err := parseFlags()
	if err != nil {
		panic(err)
	}

	// logLevel can be made a public variable so logging level can be changed dynamically.
	// TODO (derpsteb): enable once we are on go 1.21.
	// logLevel := new(slog.LevelVar)
	// handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	// logger := slog.New(handler)
	// logLevel.Set(flags.logLevel)

	logger := logger.New(logger.JSONLog, logger.VerbosityFromInt(flags.logLevel))

	if err := runServer(flags, logger); err != nil {
		panic(err)
	}
}

func runServer(flags cmdFlags, log *logger.Logger) error {
	log.With(zap.String("ip", flags.ip), zap.Int("port", defaultPort), zap.String("region", flags.region)).Infof("listening")

	router := router.New(flags.region, log)

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", flags.ip, defaultPort),
		Handler: http.HandlerFunc(router.Serve),
		// Disable HTTP/2. Serving HTTP/2 will cause some clients to use HTTP/2.
		// It seems like AWS S3 does not support HTTP/2.
		// Having HTTP/2 enabled will at least cause the aws-sdk-go V1 copy-object operation to fail.
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}

	// i.e. if TLS is enabled.
	if !flags.noTLS {
		cert, err := tls.LoadX509KeyPair(flags.certLocation+"/s3proxy.crt", flags.certLocation+"/s3proxy.key")
		if err != nil {
			return fmt.Errorf("loading TLS certificate: %w", err)
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		// TLSConfig is populated, so we can safely pass empty strings to ListenAndServeTLS.
		return server.ListenAndServeTLS("", "")
	}

	log.Warnf("TLS is disabled")
	return server.ListenAndServe()
}

func parseFlags() (cmdFlags, error) {
	noTLS := flag.Bool("no-tls", false, "disable TLS and listen on port 80, otherwise listen on 443")
	ip := flag.String("ip", defaultIP, "ip to listen on")
	region := flag.String("region", defaultRegion, "AWS region in which target bucket is located")
	certLocation := flag.String("cert", defaultCertLocation, "location of TLS certificate")
	level := flag.Int("level", defaultLogLevel, "log level")

	flag.Parse()

	netIP := net.ParseIP(*ip)
	if netIP == nil {
		return cmdFlags{}, fmt.Errorf("not a valid IPv4 address: %s", *ip)
	}

	// TODO(derpsteb): enable once we are on go 1.21.
	// logLevel := new(slog.Level)
	// if err := logLevel.UnmarshalText([]byte(*level)); err != nil {
	// 	return cmdFlags{}, fmt.Errorf("parsing log level: %w", err)
	// }

	return cmdFlags{noTLS: *noTLS, ip: netIP.String(), region: *region, certLocation: *certLocation, logLevel: *level}, nil
}

type cmdFlags struct {
	noTLS        bool
	ip           string
	region       string
	certLocation string
	// TODO(derpsteb): enable once we are on go 1.21.
	// logLevel slog.Level
	logLevel int
}
