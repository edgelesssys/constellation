//go:build integration

package test

import (
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"
)

var (
	runAwsStorage      = flag.Bool("awsStorage", false, "set to run AWS S3 Bucket Storage test")
	runAwsKms          = flag.Bool("awsKms", false, "set to run AWS KMS test")
	azConnectionString = flag.String("azStorageConn", "", "connection string for Azure storage account. Required for Azure storage test.")
	runAzStorage       = flag.Bool("azStorage", false, "set to run Azure Storage test")
	runAzKms           = flag.Bool("azKms", false, "set to run Azure KMS test")
	runAzHsm           = flag.Bool("azHsm", false, "set to run Azure HSM test")
	runGcpKms          = flag.Bool("gcpKms", false, "set to run Google KMS test")
	runGcpStorage      = flag.Bool("gcpStorage", false, "set to run Google Storage test")
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	flag.Parse()
	os.Exit(m.Run())
}

func addSuffix(s string) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return s + "-" + string(b)
}
