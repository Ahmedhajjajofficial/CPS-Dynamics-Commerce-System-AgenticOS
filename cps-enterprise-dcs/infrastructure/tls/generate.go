package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"go.uber.org/zap"
)

// CertificateConfig holds certificate generation parameters
type CertificateConfig struct {
	CommonName         string
	Organization       string
	Country            string
	ValidityYears      int
	KeySize            int
	IsCA               bool
	DNSNames           []string
	IPAddresses        []string
}

// GenerateCertificate generates a new certificate and key pair
func GenerateCertificate(cfg *CertificateConfig, logger *zap.Logger) ([]byte, []byte, error) {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, cfg.KeySize)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:         cfg.CommonName,
			Organization:       []string{cfg.Organization},
			Country:            []string{cfg.Country},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(cfg.ValidityYears, 0, 0),
		KeyUsage: 0,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		DNSNames:    cfg.DNSNames,
		IPAddresses: cfg.IPAddresses,
	}

	if cfg.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	logger.Info("Certificate generated",
		zap.String("common_name", cfg.CommonName),
		zap.Bool("is_ca", cfg.IsCA),
	)

	return certPEM, keyPEM, nil
}

// SaveCertificate saves certificate and key to files
func SaveCertificate(certPEM, keyPEM []byte, certPath, keyPath string) error {
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return err
	}
	return nil
}
