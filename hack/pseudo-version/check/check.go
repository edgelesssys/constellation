/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bazelbuild/rules_go/go/runfiles"
)

const (
	darwinArm64Filename = "pseudo_version_darwin_arm64"
	darwinAmd64Filename = "pseudo_version_darwin_amd64"
	linuxArm64Filename  = "pseudo_version_linux_arm64"
	linuxAmd64Filename  = "pseudo_version_linux_amd64"
	bucket              = "cdn-constellation-backend"
	keyPrefix           = "constellation/cas/sha256/"
)

func main() {
	checker, err := newChecker()
	if err != nil {
		log.Fatalf("failed to create checker: %v", err)
	}

	if err := checker.checkAll(); err != nil {
		log.Fatalf("failed to check pseudo-version tools: %v", err)
	}

	log.Println("All pseudo-version tools are up-to-date")
}

// a checker checks if the pseudo-version tool with the specified hash exists in S3.
type checker struct {
	files                      *runfiles.Runfiles
	downloader                 *s3manager.Downloader
	uploader                   *s3manager.Uploader
	pseudoVersionToolFilenames []string
}

// newChecker creates a new checker.
func newChecker() (*checker, error) {
	files, err := runfiles.New()
	if err != nil {
		return nil, fmt.Errorf("Failed to create runfiles: %v", err)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	}))

	return &checker{
		files:      files,
		downloader: s3manager.NewDownloader(sess),
		uploader:   s3manager.NewUploader(sess),
		pseudoVersionToolFilenames: []string{
			darwinArm64Filename,
			darwinAmd64Filename,
			linuxArm64Filename,
			linuxAmd64Filename,
		},
	}, nil
}

// checkAll checks all embedded pseudo-version tools.
func (c *checker) checkAll() error {
	for _, filename := range c.pseudoVersionToolFilenames {
		if err := c.check(filename); err != nil {
			return fmt.Errorf("failed to check pseudo-version tool (%s): %v", filename, err)
		}
	}
	return nil
}

// check checks if the pseudo-version tool with the specified hash exists in S3 and
// uploads it if it doesn't.
func (c *checker) check(filename string) error {
	log.Println("Checking pseudo-version tool:", filename)
	hash, err := c.hashPseudoVersionTool(filename)
	if err != nil {
		return fmt.Errorf("failed to hash pseudo-version tool (%s): %v", filename, err)
	}
	log.Printf("Hash: %x\n", hash)

	exists, err := c.matchesS3Hash(filename, hash)
	if err != nil {
		return fmt.Errorf("failed to check if pseudo-version tool (%s) exists in S3: %v", filename, err)
	}
	log.Println("Exists in S3:", exists)

	if !exists {
		log.Println("Uploading pseudo-version tool:", filename)
		if err := c.uploadToS3(filename, hash); err != nil {
			return fmt.Errorf("failed to upload pseudo-version tool (%s) to S3: %v", filename, err)
		}
	}

	return nil
}

// uploadToS3 uploads the pseudo-version tool with the specified hash to S3.
func (c *checker) uploadToS3(filename string, hash [32]byte) error {
	contents, err := c.files.ReadFile(fmt.Sprintf("__main__/hack/pseudo-version/%s", filename))
	if err != nil {
		return fmt.Errorf("failed to read pseudo-version tool (%s): %v", filename, err)
	}

	key := keyPrefix + fmt.Sprintf("%x", hash)
	_, err = c.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(contents),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %x to S3: %v", filename, err)
	}

	return nil
}

// matchesS3Hash checks the pseudo-version tool with the specified hash exists in S3.
func (c *checker) matchesS3Hash(filename string, hash [32]byte) (bool, error) {
	tmpfileName := filename + "-tmp"
	tmpfile, err := os.Create(tmpfileName)
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file %s: %v", tmpfileName, err)
	}
	defer os.Remove(tmpfileName)

	key := keyPrefix + fmt.Sprintf("%x", hash)
	_, err = c.downloader.Download(tmpfile, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNoSuchKeyErr(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to download %x from S3: %v", filename, err)
	}

	// A file with the hash exists in S3
	tmpfile.Close()
	return true, nil
}

// hashPseudoVersionTool hashes the specified embedded pseudo-version tool.
func (c *checker) hashPseudoVersionTool(filename string) ([32]byte, error) {
	contents, err := c.files.ReadFile(fmt.Sprintf("__main__/hack/pseudo-version/%s", filename))
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to read pseudo-version tool (%s): %v", filename, err)
	}

	return sha256.Sum256(contents), nil
}

func isNoSuchKeyErr(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() == s3.ErrCodeNoSuchKey {
			return true
		}
	}
	return false
}
