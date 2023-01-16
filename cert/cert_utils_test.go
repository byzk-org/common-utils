package cert

import (
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
	"testing"
	"time"
)

func TestCreateSm2Cert(t *testing.T) {
	caCertResult, err := CreateSm2Cert(GetCaCertTemplate(&pkix.Name{
		CommonName:         "测试Ca证书",
		Organization:       []string{"byzk"},
		OrganizationalUnit: []string{"byzk"},
		Province:           []string{"BeiJing"},
		Country:            []string{"CN"},
		Locality:           []string{"HaiDian"},
	}, time.Now().AddDate(100, 0, 0)))
	if err != nil {
		t.Error(err.Error())
		return
	}

	signCertResult, err := CreateSm2CertWithCa(GetSignCertTemplate(&pkix.Name{
		CommonName:         "localhost",
		Organization:       []string{"byzk"},
		OrganizationalUnit: []string{"byzk"},
		Province:           []string{"BeiJing"},
		Country:            []string{"CN"},
		Locality:           []string{"HaiDian"},
	}, time.Now().AddDate(100, 0, 0)), caCertResult.Cert, caCertResult.Pri)
	if err != nil {
		t.Error(err.Error())
		return
	}

	envCertResult, err := CreateSm2CertWithCa(GetEnvCertTemplate(&pkix.Name{
		CommonName:         "localhost",
		Organization:       []string{"byzk"},
		OrganizationalUnit: []string{"byzk"},
		Province:           []string{"BeiJing"},
		Country:            []string{"CN"},
		Locality:           []string{"HaiDian"},
	}, time.Now().AddDate(100, 0, 0)), caCertResult.Cert, caCertResult.Pri)
	if err != nil {
		t.Error(err.Error())
	}

	userCertResult, err := CreateSm2CertWithCa(GetUserCertTemplate(&pkix.Name{
		CommonName:         "测试用户证书",
		Organization:       []string{"byzk"},
		OrganizationalUnit: []string{"byzk"},
		Province:           []string{"BeiJing"},
		Country:            []string{"CN"},
		Locality:           []string{"HaiDian"},
	}, time.Now().AddDate(100, 0, 0)), caCertResult.Cert, caCertResult.Pri)

	if err != nil {
		t.Error(err.Error())
		return
	}

	syncCertResult, err := CreateSm2CertWithCa(GetUserCertTemplate(&pkix.Name{
		CommonName:         "测试同步证书",
		Organization:       []string{"byzk"},
		OrganizationalUnit: []string{"byzk"},
		Province:           []string{"BeiJing"},
		Country:            []string{"CN"},
		Locality:           []string{"HaiDian"},
	}, time.Now().AddDate(100, 0, 0)), caCertResult.Cert, caCertResult.Pri)
	if err != nil {
		t.Error(err.Error())
		return
	}

	printCertAndPri("CA证书", caCertResult.CertPem, caCertResult.Pri, t)
	printCertAndPri("签名证书", signCertResult.CertPem, signCertResult.Pri, t)
	printCertAndPri("加密证书", envCertResult.CertPem, envCertResult.Pri, t)
	printCertAndPri("用户证书", userCertResult.CertPem, userCertResult.Pri, t)
	printCertAndPri("同步证书", syncCertResult.CertPem, syncCertResult.Pri, t)
}

func printCertAndPri(prefix string, certPem string, key *sm2.PrivateKey, t *testing.T) {
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("============================================")
	fmt.Println(prefix + "证书信息")
	fmt.Println(certPem)
	fmt.Println()
	fmt.Println()
	privateKey, err := x509.MarshalSm2UnecryptedPrivateKey(key)
	if err != nil {
		t.Error(err.Error())
	}
	memory := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	})
	fmt.Println(string(memory))
	fmt.Println()
	fmt.Println()
	fmt.Println("============================================")
}
