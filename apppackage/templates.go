package apppackage

const (
	startRunnerTemplate = `package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/tjfoc/gmsm/gmtls"
	"github.com/tjfoc/gmsm/x509"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	//go:embed app.content
	execContent []byte
	//go:embed app.info
	exeInfoContent string
	//go:embed cert.info
	certInfoContent []byte
)

var (
	tmpMsg   []byte
	splitMsg = []byte("&&")
)

func main() {
	defer func() { recover() }()

	if len(os.Args) > 2 {
		fmt.Println("调用参数错误")
		os.Exit(2)
	}

	index := strings.Index(exeInfoContent, ";")
	if index == -1 {
		fmt.Println("未找到服务端口")
		os.Exit(2)
	}

	portStr := exeInfoContent[:index]
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		fmt.Println("转换端口信息失败")
		os.Exit(3)
	}

	if len(os.Args) == 2 {
		if os.Args[1] != "install" {
			fmt.Println("调用参数错误")
			os.Exit(4)
		}

		installApp(port)

		return
	}

	otherContent := exeInfoContent[index+1:]
	hexContent := strings.Split(otherContent, ";")
	if len(hexContent) != 2 {
		fmt.Println("获取app内容失败")
		os.Exit(9)
	}

	hexAppContent := hexContent[0]
	hexAppVersionContent := hexContent[1]
	appContentBytes, err := hex.DecodeString(hexAppContent)
	if err != nil {
		fmt.Println("解析应用信息失败")
		os.Exit(9)
	}

	appVersionContentBytes, err := hex.DecodeString(hexAppVersionContent)
	if err != nil {
		fmt.Println("解析应用版本信息失败")
		os.Exit(9)
	}

	fmt.Printf("应用信息：\n%s\n\n", appContentBytes)
	fmt.Printf("应用版本信息：\n%s\n", appVersionContentBytes)

}

func installApp(port int64) {
	conn := GetClientConn(port)
	msgChannel := make(chan []byte)
	go func() {
		if err := readData(conn, msgChannel); err != nil {
			close(msgChannel)
		}
	}()

	writeDataStr(conn, hex.EncodeToString([]byte("import"))+"&&")
	readMsg(msgChannel)
	md5Sum := md5.Sum(execContent)
	sha1Sum := sha1.Sum(execContent)

	writeDataStr(conn, hex.EncodeToString([]byte(hex.EncodeToString(md5Sum[:]))))
	writeDataStr(conn, "&&")

	writeDataStr(conn, hex.EncodeToString([]byte(hex.EncodeToString(sha1Sum[:]))))
	writeDataStr(conn, "&&")

	tmpDir, err := ioutil.TempDir("", "appRunnerPlatform*")
	if err != nil {
		fmt.Println("创建文件临时缓存目录失败")
		os.Exit(9)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "appRunner")
	if err = ioutil.WriteFile(tmpFile, execContent, 0666); err != nil {
		fmt.Println("写入临时文件失败")
		os.Exit(9)
	}

	writeDataStr(conn, hex.EncodeToString([]byte(tmpFile)))
	writeDataStr(conn, "&&")

	readMsg(msgChannel)
	writeDataStr(conn, "end!!&&")

	fmt.Println("导入应用成功")

}

func GetClientConn(port int64) *gmtls.Conn {
	certInfoMap := make(map[string]string)
	err := json.Unmarshal(certInfoContent, &certInfoMap)
	if err != nil {
		fmt.Println("读取证书内容失败")
		os.Exit(5)
	}

	userCertPem, userCertOk := certInfoMap["u"]
	userKeyPem, userKeyOk := certInfoMap["uk"]
	rootCertPem, rootCertPemOk := certInfoMap["r"]
	if !userCertOk || !userKeyOk || !rootCertPemOk {
		fmt.Println("读取用户凭证失败")
		os.Exit(6)
	}

	if !validCertExpire([]byte(userCertPem)) {
		fmt.Println("用户凭证已过期")
		os.Exit(7)
	}

	userCert, err := gmtls.GMX509KeyPairsSingle([]byte(userCertPem), []byte(userKeyPem))
	if err != nil {
		fmt.Println("解析用户凭证失败")
		os.Exit(7)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(rootCertPem))

	conn, err := gmtls.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), &gmtls.Config{
		GMSupport:    &gmtls.GMSupport{},
		ServerName:   "localhost",
		Certificates: []gmtls.Certificate{userCert},
		RootCAs:      certPool,
		ClientAuth:   gmtls.RequireAndVerifyClientCert,
	})
	if err != nil {
		fmt.Println("访问本地程序失败!")
		os.Exit(12)
	}

	return conn
}

func validCertExpire(certPemBytes []byte) bool {
	block, _ := pem.Decode(certPemBytes)
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	return time.Now().Before(certificate.NotAfter)
}

func readData(conn net.Conn, msgChannel chan<- []byte) (returnErr error) {
	defer func() {
		e := recover()
		if e != nil {
			returnErr = errors.New("读取数据出现异常")
		}
	}()
	buffer := make([]byte, 1024*1024)
	for {
		read, err := conn.Read(buffer)
		if err != nil {
			return err
		}

		tmpByte := make([]byte, read)
		copy(tmpByte, buffer[:read])
		if tmpMsg != nil {
			tmpByte = bytes.Join([][]byte{
				tmpMsg,
				tmpByte,
			}, nil)
			tmpMsg = nil
		}

		if !bytes.Contains(tmpByte, splitMsg) {
			if tmpMsg != nil {
				tmpMsg = make([]byte, len(tmpByte))
			}
			tmpMsg = append(tmpMsg, tmpByte...)
			continue
		}
		splitByte := bytes.Split(tmpByte, splitMsg)
		msgChannel <- splitByte[0]
		if len(splitByte) > 2 {
			for i := 1; i < len(splitByte)-1; i++ {
				msgChannel <- splitByte[i]
			}
		}

		tmpOtherMsg := splitByte[len(splitByte)-1]
		if len(tmpOtherMsg) == 0 {
			tmpMsg = nil
		} else {
			if tmpMsg == nil {
				tmpMsg = make([]byte, len(tmpByte))
			}
			tmpMsg = append(tmpMsg, tmpOtherMsg...)
		}

	}
}

func readMsg(msgChannel chan []byte) string {
	msgByte, isOpen := <-msgChannel
	if !isOpen || len(msgByte) == 0 {
		fmt.Println("未知的处理异常")
		os.Exit(8)
	}

	msgStr := string(msgByte)
	decodeString, err := hex.DecodeString(msgStr)
	if err != nil {
		fmt.Println("解析数据结构失败")
		os.Exit(8)
	}
	if string(decodeString) == "error" {
		msgByte, isOpen = <-msgChannel
		if !isOpen || len(msgByte) == 0 {
			fmt.Println("未知的处理异常")
			os.Exit(8)
		}
		msgStrByte, err := hex.DecodeString(string(msgByte))
		if err != nil {
			fmt.Println("解析数据结构失败")
			os.Exit(8)
		}
		fmt.Println(string(msgStrByte))
		os.Exit(8)
	}

	msgByte, isOpen = <-msgChannel
	if !isOpen || len(msgByte) == 0 {
		fmt.Println("未知的处理异常")
		os.Exit(8)
	}

	return string(msgByte)
}

func writeDataStr(conn net.Conn, data string) {
	writeData(conn, []byte(data))
}

func writeData(conn net.Conn, data []byte) {
	_, err := conn.Write(data)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("发送数据失败")
		os.Exit(8)
	}
}
`
)
