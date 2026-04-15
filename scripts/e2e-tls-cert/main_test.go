package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseHostsTrimsAndSkipsEmptyEntries(t *testing.T) {
	t.Parallel()

	hosts := parseHosts(" 127.0.0.1 , localhost,, example.test , ")
	if len(hosts) != 3 {
		t.Fatalf("expected 3 hosts, got %d: %#v", len(hosts), hosts)
	}
	if hosts[0] != "127.0.0.1" || hosts[1] != "localhost" || hosts[2] != "example.test" {
		t.Fatalf("unexpected parsed hosts: %#v", hosts)
	}
}

func TestParseHostsReturnsEmptySliceForBlankInput(t *testing.T) {
	t.Parallel()

	if hosts := parseHosts(" , , "); len(hosts) != 0 {
		t.Fatalf("expected no hosts for blank input, got %#v", hosts)
	}
}

func TestEnsureParentDirCreatesMissingParent(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "nested", "tls", "localhost.pem")
	if err := ensureParentDir(target); err != nil {
		t.Fatalf("ensureParentDir returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(target)); err != nil {
		t.Fatalf("expected parent directory to exist, got %v", err)
	}
}

func TestEnsureParentDirAllowsBareFilename(t *testing.T) {
	t.Parallel()

	if err := ensureParentDir("localhost.pem"); err != nil {
		t.Fatalf("expected bare filename to be accepted, got %v", err)
	}
}

func TestGenerateCertificateIncludesProvidedHosts(t *testing.T) {
	t.Parallel()

	certificatePEM, privateKeyPEM, err := generateCertificate([]string{"127.0.0.1", "localhost", "example.test"})
	if err != nil {
		t.Fatalf("generateCertificate returned error: %v", err)
	}

	certificate := mustParseCertificate(t, certificatePEM)
	privateKey := mustParseRSAPrivateKey(t, privateKeyPEM)

	assertCertificateSubjectAndSANs(t, certificate)
	assertCertificateLifetime(t, certificate)
	assertCertificateMatchesPrivateKey(t, certificate, privateKey)
}

func mustParseCertificate(t *testing.T, certificatePEM []byte) *x509.Certificate {
	t.Helper()

	certBlock, _ := pem.Decode(certificatePEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		t.Fatalf("expected certificate PEM block, got %#v", certBlock)
	}
	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return certificate
}

func mustParseRSAPrivateKey(t *testing.T, privateKeyPEM []byte) *rsa.PrivateKey {
	t.Helper()

	keyBlock, _ := pem.Decode(privateKeyPEM)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		t.Fatalf("expected rsa private key PEM block, got %#v", keyBlock)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	return privateKey
}

func assertCertificateSubjectAndSANs(t *testing.T, certificate *x509.Certificate) {
	t.Helper()

	if certificate.Subject.CommonName != "ovumcy-e2e-localhost" {
		t.Fatalf("unexpected certificate common name %q", certificate.Subject.CommonName)
	}
	if !containsIP(certificate.IPAddresses, net.ParseIP("127.0.0.1")) {
		t.Fatalf("expected localhost IP SAN, got %#v", certificate.IPAddresses)
	}
	if !containsString(certificate.DNSNames, "localhost") || !containsString(certificate.DNSNames, "example.test") {
		t.Fatalf("expected DNS SANs to include localhost and example.test, got %#v", certificate.DNSNames)
	}
}

func assertCertificateLifetime(t *testing.T, certificate *x509.Certificate) {
	t.Helper()

	if !certificate.NotAfter.After(certificate.NotBefore) {
		t.Fatalf("expected certificate lifetime to be positive, got notBefore=%s notAfter=%s", certificate.NotBefore, certificate.NotAfter)
	}
	if lifetime := certificate.NotAfter.Sub(certificate.NotBefore); lifetime < 23*time.Hour {
		t.Fatalf("expected certificate lifetime near %s, got %s", time.Duration(certificateHours)*time.Hour, lifetime)
	}
}

func assertCertificateMatchesPrivateKey(t *testing.T, certificate *x509.Certificate, privateKey *rsa.PrivateKey) {
	t.Helper()

	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("expected RSA public key, got %T", certificate.PublicKey)
	}
	if publicKey.N.Cmp(privateKey.N) != 0 {
		t.Fatal("expected certificate public key to match generated private key")
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsIP(values []net.IP, target net.IP) bool {
	for _, value := range values {
		if value.Equal(target) {
			return true
		}
	}
	return false
}
