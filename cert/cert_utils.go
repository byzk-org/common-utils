package cert

import (
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
	"math/big"
	"sync"
	"time"
)

var lock sync.Mutex

// CreateSm2Cert 创建sm2证书
func CreateSm2Cert(certInfo *x509.Certificate) (*Sm2CertCreateResult, error) {
	return CreateSm2CertWithCa(certInfo, nil, nil)
}

// CreateSm2CertWithCa 创建sm2证书伴随ca证书信息
func CreateSm2CertWithCa(certInfo, caCert *x509.Certificate, caPrivate *sm2.PrivateKey) (*Sm2CertCreateResult, error) {
	lock.Lock()
	defer lock.Unlock()

	if certInfo == nil {
		return nil, errors.New("获取要创建的证书信息失败")
	}

	certInfo.SerialNumber = big.NewInt(time.Now().UnixNano())

	key, err := sm2.GenerateKey(nil)
	if err != nil {
		return nil, errors.New("创建公钥失败 => " + err.Error())
	}

	if caCert == nil {
		caCert = certInfo
	}

	haveCa := true

	if caPrivate == nil {
		caPrivate = key
		haveCa = false
	}

	certPem, err := x509.CreateCertificateToPem(certInfo, caCert, key.Public().(*sm2.PublicKey), caPrivate)
	if err != nil {
		return nil, errors.New("创建证书失败 => " + err.Error())
	}

	block, _ := pem.Decode(certPem)
	parseCertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("创建证书失败 => " + err.Error())
	}

	if haveCa {
		if err = parseCertificate.CheckSignatureFrom(caCert); err != nil {
			return nil, errors.New("证书签名验证失败")
		}
	} else {
		if err = parseCertificate.CheckSignatureFrom(parseCertificate); err != nil {
			return nil, errors.New("证书签名验证失败")
		}
	}

	privateKey, err := x509.MarshalSm2UnecryptedPrivateKey(key)
	if err != nil {
		return nil, errors.New("转换私钥到pem失败")
	}
	memory := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	})
	return &Sm2CertCreateResult{
		Cert:       parseCertificate,
		Pri:        key,
		PriPem:     string(memory),
		PriPemDer:  memory,
		CertDer:    block.Bytes,
		CertPem:    string(certPem),
		CertPemDer: certPem,
	}, nil
}

// GetCaCertTemplate 获取Ca证书模板
func GetCaCertTemplate(subject *pkix.Name, expire time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject:               *subject,
		SignatureAlgorithm:    x509.SM2WithSM3,
		NotBefore:             time.Now(),
		NotAfter:              expire,
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
}

// GetSignCertTemplate 获取签名证书模板
func GetSignCertTemplate(subject *pkix.Name, expire time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject:               *subject,
		SignatureAlgorithm:    x509.SM2WithSM3,
		NotBefore:             time.Now(),
		NotAfter:              expire,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
	}
}

// GetEnvCertTemplate 获取加密证书模板
func GetEnvCertTemplate(subject *pkix.Name, expire time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject:               *subject,
		SignatureAlgorithm:    x509.SM2WithSM3,
		NotBefore:             time.Now(),
		NotAfter:              expire,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment,
	}
}

// GetUserCertTemplate 获取用户证书模板
func GetUserCertTemplate(subject *pkix.Name, expire time.Time) *x509.Certificate {
	return &x509.Certificate{
		Subject:               *subject,
		SignatureAlgorithm:    x509.SM2WithSM3,
		NotBefore:             time.Now(),
		NotAfter:              expire,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageContentCommitment | x509.KeyUsageKeyEncipherment,
	}
}

type Sm2CertCreateResult struct {
	Cert       *x509.Certificate
	Pri        *sm2.PrivateKey
	PriPem     string
	PriPemDer  []byte
	CertDer    []byte
	CertPem    string
	CertPemDer []byte
}
