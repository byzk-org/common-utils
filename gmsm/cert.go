package gmsm

import (
	"encoding/pem"
	"errors"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

func ParsePubKeyByCertPem(certPem string) (*sm2.PublicKey, error) {
	decode, _ := pem.Decode([]byte(certPem))
	if decode == nil {
		return nil, errors.New("解析证书失败")
	}

	certificate, err := x509.ParseCertificate(decode.Bytes)
	if err != nil {
		return nil, errors.New("转换证书失败")
	}

	key, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return nil, errors.New("转换公钥失败")
	}

	pubKey, err := x509.ParseSm2PublicKey(key)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}
