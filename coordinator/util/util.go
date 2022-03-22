package util

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"net"

	"golang.org/x/crypto/hkdf"
)

// DeriveKey derives a key from a secret.
//
// TODO: decide on a secure key derivation function.
func DeriveKey(secret, salt, info []byte, length uint) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateCertificateSerialNumber generates a random serial number for an X.509 certificate.
func GenerateCertificateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

// GenerateRandomBytes reads length bytes from getrandom(2) if available, /dev/urandom otherwise.
func GenerateRandomBytes(length int) ([]byte, error) {
	nonce := make([]byte, length)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func GetIPAddr() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func GetInterfaceIP(netInterface string) (string, error) {
	netif, err := net.InterfaceByName(netInterface)
	if err != nil {
		return "", fmt.Errorf("could not find interface %s: %w", netInterface, err)
	}
	addrs, err := netif.Addrs()
	if err != nil {
		return "", fmt.Errorf("could not retrieve interface ip addresses %s: %w", netInterface, err)
	}
	for _, addr := range addrs {
		if ipn, ok := addr.(*net.IPNet); ok {
			if ip := ipn.IP.To4(); ip != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("interface %s don't have an ipv4 address", netInterface)
}
