// Verify verifies an MAA JWT and prints the SNP ID key digest on success.
package main

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<maa-jwt>")
		return
	}
	idKeyDigest, err := getIDKeyDigest(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully validated ID key digest:", idKeyDigest)
}

func getIDKeyDigest(rawToken string) (string, error) {
	// Parse token.
	token, err := jwt.ParseSigned(rawToken)
	if err != nil {
		return "", err
	}

	// Get JSON Web Key set.
	keySetBytes, err := httpGet("https://sharedeus.eus.attest.azure.net/certs")
	if err != nil {
		return "", err
	}
	keySet, err := parseKeySet(keySetBytes)
	if err != nil {
		return "", err
	}

	// Get claims. Private claims contain ID Key digest.

	var publicClaims jwt.Claims

	var privateClaims struct {
		IsolationTEE struct {
			IDKeyDigest string `json:"x-ms-sevsnpvm-idkeydigest"`
		} `json:"x-ms-isolation-tee"`
	}

	if err := token.Claims(&keySet, &publicClaims, &privateClaims); err != nil {
		return "", err
	}
	if err := publicClaims.Validate(jwt.Expected{Time: time.Now()}); err != nil {
		return "", err
	}

	return privateClaims.IsolationTEE.IDKeyDigest, nil
}

func parseKeySet(keySetBytes []byte) (jose.JSONWebKeySet, error) {
	var rawKeySet struct {
		Keys []struct {
			X5c []string
			Kid string
		}
	}
	if err := json.Unmarshal(keySetBytes, &rawKeySet); err != nil {
		return jose.JSONWebKeySet{}, err
	}

	var keySet jose.JSONWebKeySet
	for _, key := range rawKeySet.Keys {
		rawCert, _ := base64.StdEncoding.DecodeString(key.X5c[0])
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return jose.JSONWebKeySet{}, err
		}
		keySet.Keys = append(keySet.Keys, jose.JSONWebKey{KeyID: key.Kid, Key: cert.PublicKey})
	}

	return keySet, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
