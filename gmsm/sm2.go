package gmsm

import (
	"crypto/rand"
	"errors"
	"github.com/tjfoc/gmsm/sm2"
)

// Sm2Encrypt Sm2加密
func Sm2Encrypt(pubKey *sm2.PublicKey, data []byte) ([]byte, error) {
	encrypt, err := sm2.Encrypt(pubKey, data, rand.Reader)
	if err != nil {
		return nil, errors.New("加密数据失败")
	}
	return encrypt, err
}

// Sm2Decrypt sn2解密
func Sm2Decrypt(pri *sm2.PrivateKey, data []byte) ([]byte, error) {
	decrypt, err := sm2.Decrypt(pri, data)
	if err != nil {
		return nil, errors.New("解密数据失败")
	}
	return decrypt, nil
}
