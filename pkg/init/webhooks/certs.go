package webhooks

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type generatedCert struct {
	privateKey []byte
	publicKey  []byte
	expiresAt  time.Time
}

func generateCert(webhookService types.NamespacedName, additionalDNSNames []string) (*generatedCert, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}

	dnsNames := getServiceCertDNSNames(webhookService)
	dnsNames = append(dnsNames, additionalDNSNames...)
	expiresAt := time.Now().Add(10 * 365 * 24 * time.Hour).UTC() // 10 years
	cert := x509.Certificate{
		Subject: pkix.Name{
			CommonName: dnsNames[0],
		},
		DNSNames:     dnsNames,
		NotBefore:    time.Now().UTC(),
		NotAfter:     expiresAt,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		SerialNumber: serial,
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	encodedKey, err := encodePrivateKey(key)
	if err != nil {
		return nil, err
	}

	encodedCert, err := encodeCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	return &generatedCert{
		privateKey: encodedKey,
		publicKey:  encodedCert,
		expiresAt:  expiresAt,
	}, nil
}

func getServiceCertDNSNames(webhookService types.NamespacedName) []string {
	return []string{
		fmt.Sprintf("%s.%s.svc", webhookService.Name, webhookService.Namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", webhookService.Name, webhookService.Namespace),
	}
}

func encodePrivateKey(key any) ([]byte, error) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	keyPEM := &bytes.Buffer{}
	err = pem.Encode(keyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})
	return keyPEM.Bytes(), err
}

func encodeCertificate(certBytes []byte) ([]byte, error) {
	certPEM := &bytes.Buffer{}
	err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	return certPEM.Bytes(), err
}
