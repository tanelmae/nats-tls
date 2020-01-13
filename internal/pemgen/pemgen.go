package pemgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/tanelmae/nats-tls/internal/config"
	"time"
)

const (
	pemPrivateType = "RSA PRIVATE KEY"
	pemCertType    = "CERTIFICATE"
)

// CA struct holds CA key and certificate data
type CA struct {
	key  *rsa.PrivateKey
	cert x509.Certificate
}

// GenCA is for creatin CA key and certificate
func GenCA(conf config.CertConfig) (*CA, error) {
	log.Println("Creating CA key and certificate")
	key, err := rsa.GenerateKey(rand.Reader, conf.KeyLength)
	if err != nil {
		return nil, err
	}

	cert := x509.Certificate{
		SerialNumber: genCertSerial(),
		Subject: pkix.Name{
			CommonName:   conf.Subject.CN,
			Organization: []string{conf.Subject.Org},
			Country:      []string{conf.Subject.Country},
		},
		SubjectKeyId:          genSubjectKeyID(key),
		NotBefore:             time.Now(),
		NotAfter:              conf.TTL.Expiration,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	log.Printf("Certificate valid untiL: %s", conf.TTL.Expiration.Format("2006-01-02 15:04:05"))
	caBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	writeKey(key, conf.Path, conf.Name)
	writeCert(caBytes, conf.Path, conf.Name)
	log.Printf("CA key and cert created at %s", conf.Path)
	return &CA{key: key, cert: cert}, nil
}

// GenSignedCerts is used for creating signed ceritifcates for routes, servers and clients
func GenSignedCerts(conf config.CertConfig, ca CA) error {
	log.Printf("Creating '%s' key and certificate", conf.Name)
	key, err := rsa.GenerateKey(rand.Reader, conf.KeyLength)
	if err != nil {
		return err
	}

	cert := &x509.Certificate{
		SerialNumber: genCertSerial(),
		Subject: pkix.Name{
			CommonName:   conf.Subject.CN,
			Organization: []string{conf.Subject.Org},
			Country:      []string{conf.Subject.Country},
		},
		SubjectKeyId:          genSubjectKeyID(key),
		AuthorityKeyId:        ca.cert.SubjectKeyId,
		NotBefore:             time.Now(),
		NotAfter:              conf.TTL.Expiration,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames:              conf.DNS,
		IsCA:                  false,
		BasicConstraintsValid: true,
	}

	log.Printf("Certificate valid untiL: %s", conf.TTL.Expiration.Format("2006-01-02 15:04:05"))
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, &ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		return err
	}

	writeKey(key, conf.Path, conf.Name)
	writeCert(certBytes, conf.Path, conf.Name)
	log.Printf("Created at %s", conf.Path)
	return nil
}

func genSubjectKeyID(key *rsa.PrivateKey) []byte {
	h := sha1.New()
	h.Write(key.PublicKey.N.Bytes())
	return h.Sum(nil)
}

func genCertSerial() *big.Int {
	return big.NewInt(time.Now().UnixNano())
}

func writeCert(cert []byte, path, name string) error {
	os.MkdirAll(path, 0755)
	pemCrt, err := os.Create(fmt.Sprintf("%s/%s.pem", path, name))
	if err != nil {
		return err
	}
	defer pemCrt.Close()
	pem.Encode(pemCrt, &pem.Block{
		Type:  pemCertType,
		Bytes: cert,
	})
	return nil
}

func writeKey(key *rsa.PrivateKey, path, name string) error {
	os.MkdirAll(path, 0755)
	pemKey, err := os.Create(fmt.Sprintf("%s/%s-key.pem", path, name))
	if err != nil {
		return err
	}
	defer pemKey.Close()
	pem.Encode(pemKey, &pem.Block{
		Type:  pemPrivateType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return nil
}
