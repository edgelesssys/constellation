package main

import (
	"context"
	"encoding/base64"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ip   = flag.String("ip", "", "IP of the node to push the key to (required)")
	port = flag.String("port", "9000", "Port of the node to push the key to (required)")
	key  = flag.String("key", "QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUE=", "base64 encoded state disk key")
)

func main() {
	flag.Parse()

	if *ip == "" || *port == "" || *key == "" {
		flag.Usage()
		os.Exit(1)
	}

	decryptionKey, err := base64.StdEncoding.DecodeString(*key)
	if err != nil {
		log.Fatal(err)
	}

	addr := net.JoinHostPort(*ip, *port)
	log.Printf("Pushing key %v to node at %s", decryptionKey, addr)

	clientCfg, err := atls.CreateUnverifiedClientTLSConfig()
	if err != nil {
		log.Fatal(err)
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(clientCfg)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := keyproto.NewAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := client.PushStateDiskKey(ctx, &keyproto.PushStateDiskKeyRequest{
		StateDiskKey: decryptionKey,
	}); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully pushed key to node")
}
