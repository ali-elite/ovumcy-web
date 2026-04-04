package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultHosts     = "127.0.0.1,localhost"
	keyBits          = 2048
	certificateHours = 24
)

func main() {
	certPath := flag.String("cert", "", "absolute or relative path for the generated certificate PEM")
	keyPath := flag.String("key", "", "absolute or relative path for the generated RSA private key PEM")
	hostsFlag := flag.String("hosts", defaultHosts, "comma-separated DNS names or IPs to include in the certificate")
	flag.Parse()

	if *certPath == "" || *keyPath == "" {
		exitf("both --cert and --key are required")
	}

	hosts := parseHosts(*hostsFlag)
	if len(hosts) == 0 {
		exitf("at least one host is required")
	}

	certificatePEM, privateKeyPEM, err := generateCertificate(hosts)
	if err != nil {
		exitf("generate certificate: %v", err)
	}

	if err := ensureParentDir(*certPath); err != nil {
		exitf("create certificate directory: %v", err)
	}
	if err := ensureParentDir(*keyPath); err != nil {
		exitf("create key directory: %v", err)
	}

	if err := os.WriteFile(*certPath, certificatePEM, 0o644); err != nil {
		exitf("write certificate: %v", err)
	}
	if err := os.WriteFile(*keyPath, privateKeyPEM, 0o600); err != nil {
		exitf("write private key: %v", err)
	}
}

func generateCertificate(hosts []string) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now().Add(-time.Hour)
	notAfter := notBefore.Add(certificateHours * time.Hour)
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "ovumcy-e2e-localhost",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}

	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
			continue
		}
		template.DNSNames = append(template.DNSNames, host)
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return certificatePEM, privateKeyPEM, nil
}

func parseHosts(raw string) []string {
	parts := strings.Split(raw, ",")
	hosts := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			hosts = append(hosts, trimmed)
		}
	}
	return hosts
}

func ensureParentDir(path string) error {
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
