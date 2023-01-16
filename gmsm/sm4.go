package gmsm

import (
	"errors"
	"github.com/byzk-org/common-utils/random"
	"github.com/tjfoc/gmsm/sm4"
	"io"
	"os"
)

const (
	sm4EncDataLen = 1024 * 1024
)

// Sm4Encrypt sm4加密
func Sm4Encrypt(key, plainText []byte) ([]byte, error) {
	return sm4.Sm4Ecb(key, plainText, true)
}

// Sm4Encrypt2File 加密sm4到文件
func Sm4Encrypt2File(key []byte, srcFile, destFile *os.File) error {
	var (
		err      error
		readSize = sm4EncDataLen
		s        int
		encrypt  []byte
	)

	tmpBuffer := make([]byte, readSize)
	for {
		s, err = srcFile.Read(tmpBuffer)
		if err != nil && err == io.EOF {
			return nil
		}

		if err != nil {
			return errors.New("读取文件内容失败")
		}

		encryptContent := tmpBuffer[:s]
		encrypt, err = Sm4Encrypt(key, encryptContent)
		if err != nil {
			return errors.New("加密原文件内容失败")
		}

		_, err = destFile.Write(encrypt)
		if err != nil {
			return errors.New("写出文件失败")
		}

	}

}

// Sm4RandomKey Sm4随机ke
func Sm4RandomKey() []byte {
	return []byte(random.GetRandomString(16))[:16]
}
