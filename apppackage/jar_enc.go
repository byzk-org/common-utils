package apppackage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type encStruct struct {
	Password string   `json:"password,omitempty"`
	To       string   `json:"to,omitempty"`
	Form     string   `json:"form,omitempty"`
	Include  []string `json:"include,omitempty"`
	Exclude  []string `json:"exclude,omitempty"`
}

type encJarResult struct {
	Md5       string
	Sha1      string
	Algorithm string
	KeySize   string
	IvSize    string
	Password  string
}

// jarEncrypt jar包加密
func jarEncrypt(jdkPath, encJarPath, srcJarPath, distJarPath, password string, include, exclude []string) (map[string]string, error) {
	if jdkPath == "" {
		return nil, errors.New("获取java命令失败")
	}
	dir, err := ioutil.TempDir("", "byptE*")
	if err != nil {
		return nil, errors.New("创建临时存储目录失败")
	}
	defer os.RemoveAll(dir)

	infoJsonPath := filepath.Join(dir, "i")

	toFilePath := filepath.Join(dir, "t")

	d := &encStruct{
		Password: password,
		Include:  include,
		Exclude:  exclude,
		Form:     srcJarPath,
		To:       toFilePath,
	}

	marshal, _ := json.Marshal(d)
	if err = ioutil.WriteFile(infoJsonPath, marshal, 0666); err != nil {
		return nil, errors.New("写出打包信息失败")
	}

	command := exec.Command(jdkPath, "-jar", encJarPath, infoJsonPath)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, errors.New(string(output))
	}

	resultInfo, err := ioutil.ReadFile(filepath.Join(dir, "xjar.go"))
	if err != nil {
		return nil, errors.New("读取加密信息失败")
	}

	distFile, err := os.OpenFile(distJarPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.New("写出到目标文件失败")
	}
	defer distFile.Close()

	srcFile, err := os.OpenFile(toFilePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, errors.New("打开加密文件失败")
	}
	defer srcFile.Close()

	_, err = io.Copy(distFile, srcFile)
	if err != nil {
		return nil, errors.New("写出文件内容失败")
	}

	tmpResult := make(map[string]string)
	err = json.Unmarshal(resultInfo, &tmpResult)
	if err != nil {
		return nil, errors.New("转换加密结果失败")
	}

	for k, v := range tmpResult {
		tmpResult[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}

	return tmpResult, nil
}