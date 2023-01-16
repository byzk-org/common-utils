package apppackage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/byzk-org/common-utils/compress"
	"github.com/byzk-org/common-utils/gmsm"
	"github.com/byzk-org/common-utils/hash"
	"github.com/byzk-org/common-utils/random"
	"github.com/tjfoc/gmsm/sm2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type PackType string

func (p *PackType) String() string {
	return string(*p)
}

const (
	PackTypeInstallApp    PackType = "install-app"
	PackTypeInstallPlugin PackType = "install-plugin"
	PackTypeInstallJdk    PackType = "install-jdk"
)

type AppEnvConfig struct {
	Name       string `json:"name,omitempty"`
	Desc       string `json:"desc,omitempty"`
	DefaultVal string `json:"defaultVal,omitempty"`
}

type AppInfoInterface interface {
	GetName() string
	GetDesc() string
	GetEnvConfig() []*AppEnvConfig
	GetJdkStartArgs() []string
}

type tmpAppInfo struct {
	Name         string          `json:"name,omitempty"`
	Desc         string          `json:"desc,omitempty"`
	EnvConfig    []*AppEnvConfig `json:"envConfigInfos,omitempty"`
	JdkStartArgs []string        `json:"jdkStartArgs,omitempty"`
	OS           string          `json:"os,omitempty"`
	ARCH         string          `json:"arch,omitempty"`
}

type ZipInfo struct {
	Cmd             PackType
	JarFile         string
	RootCertPem     string
	UserCertPem     string
	UserKeyPem      string
	EncryptJarUrl   string
	Exclude         []string
	Include         []string
	LocalServerPort string
	// app扩展信息，展示的时候进行显示
	AppInfo        AppInfoInterface
	AppVersionInfo AppInfoInterface
}

func NewZipOperation(info *ZipInfo) *zipOperation {
	dir, err := ioutil.TempDir("", "byptPack*")
	if err != nil {
		return &zipOperation{Err: err}
	}

	pubKey, err := gmsm.ParsePubKeyByCertPem(info.RootCertPem)
	if err != nil {
		return &zipOperation{
			Err: errors.New("转换公钥失败"),
		}
	}

	return &zipOperation{
		zipInfo:            info,
		pluginsContentPath: make([]string, 0, 8),
		Err:                err,
		operationDir:       dir,
		pubKey:             pubKey,
	}
}

// zipOperation 压缩操作
type zipOperation struct {
	sync.Mutex
	// 加密jar内容
	//encJarContent      []byte
	encJarContentPath string
	//pluginsContent     [][]byte
	pluginsContentPath []string
	//jdkContent         []byte
	jdkContentPath string
	jarPass        interface{}
	zipInfo        *ZipInfo
	operationDir   string
	Err            error

	pubKey *sm2.PublicKey

	JdkPath    string
	EncJarPath string
}

func (z *zipOperation) SetJdkPath(jdkPath string) *zipOperation {
	z.JdkPath = jdkPath
	return z
}

func (z *zipOperation) SetEncJarPath(encJarPath string) *zipOperation {
	z.EncJarPath = encJarPath
	return z
}

// ZipApp 创建与压缩加密可执行jar
func (z *zipOperation) ZipApp() *zipOperation {
	z.Lock()
	defer z.Unlock()

	if z.Err != nil {
		return z
	}

	if z.zipInfo == nil || z.zipInfo.JarFile == "" {
		z.Err = errors.New("要打包的内容不能为空")
		return z
	}

	password := random.GetRandomString(64)

	z.encJarContentPath = filepath.Join(z.operationDir, "encJ")

	jarPass, err := jarEncrypt(z.JdkPath, z.EncJarPath, z.zipInfo.JarFile, z.encJarContentPath, password, z.zipInfo.Include, z.zipInfo.Exclude)
	if err != nil {
		z.Err = err
		return z
	}

	z.jarPass = jarPass
	return z
}

// ZipPlugin 压缩插件
func (z *zipOperation) ZipPlugin(pluginInfo string, contentPath string) *zipOperation {
	z.Lock()
	defer z.Unlock()
	contentFile, err := os.OpenFile(contentPath, os.O_RDONLY, 0666)
	if err != nil {
		z.Err = errors.New("获取插件文件失败")
		return z
	}
	defer contentFile.Close()

	pluginInfoBytes := []byte(pluginInfo)
	pluginInfoBytesLen := strconv.FormatInt(int64(len(pluginInfoBytes)), 10)

	now := time.Now().UnixNano()
	fileName := strconv.FormatInt(now, 10)
	pluginSavePath := filepath.Join(z.operationDir, fileName)
	_ = os.RemoveAll(pluginSavePath)
	pluginSaveFile, err := os.Create(pluginSavePath)
	if err != nil {
		z.Err = errors.New("创建插件保存目录失败")
		return z
	}
	defer pluginSaveFile.Close()

	pluginSaveFile.WriteString(pluginInfoBytesLen)
	pluginSaveFile.WriteString(";")
	pluginSaveFile.Write(pluginInfoBytes)
	io.Copy(pluginSaveFile, contentFile)
	z.pluginsContentPath = append(z.pluginsContentPath, pluginSavePath)

	return z
}

func (z *zipOperation) ZipJdk(jdkTarGzPath string) *zipOperation {
	z.Lock()
	defer z.Unlock()

	jdkTarGzFile, err := os.OpenFile(jdkTarGzPath, os.O_RDONLY, 0666)
	if err != nil {
		z.Err = errors.New("打开jdk原文件失败")
		return z
	}
	defer jdkTarGzFile.Close()

	z.jdkContentPath = filepath.Join(z.operationDir, ".jdk")
	_ = os.RemoveAll(z.jdkContentPath)
	jdkSaveFile, err := os.Create(z.jdkContentPath)
	if err != nil {
		z.Err = errors.New("创建jdk保存文件失败")
		return z
	}
	defer jdkSaveFile.Close()

	_, err = io.Copy(jdkSaveFile, jdkTarGzFile)
	if err != nil {
		z.Err = errors.New("复制jdk文件失败")
	}

	return z
}

// End 结束
func (z *zipOperation) End(goos, goarch string) ([]byte, error) {
	if z.Err != nil {
		return nil, z.Err
	}
	var (
		err     error
		marshal []byte
	)
	defer os.RemoveAll(z.operationDir)

	if z.zipInfo.AppInfo == nil || z.zipInfo.AppVersionInfo == nil {
		return nil, errors.New("app信息不能为空")
	}

	if z.zipInfo.Cmd == "" {
		return nil, errors.New("要操作的命令码不能为空")
	}

	appInfoBytes, _ := json.Marshal(&tmpAppInfo{
		Name: z.zipInfo.AppInfo.GetName(),
		Desc: z.zipInfo.AppInfo.GetDesc(),
	})

	appVersionBytes, _ := json.Marshal(&tmpAppInfo{
		Name:         z.zipInfo.AppVersionInfo.GetName(),
		Desc:         z.zipInfo.AppVersionInfo.GetDesc(),
		EnvConfig:    z.zipInfo.AppVersionInfo.GetEnvConfig(),
		JdkStartArgs: z.zipInfo.AppVersionInfo.GetJdkStartArgs(),
		OS:           goos,
		ARCH:         goarch,
	})

	if z.zipInfo.LocalServerPort == "" {
		return nil, errors.New("服务端口不能为空")
	}

	distContentFilePath := filepath.Join(z.operationDir, "ct")
	_ = os.RemoveAll(distContentFilePath)
	distContentFile, err := os.Create(distContentFilePath)
	if err != nil {
		z.Err = errors.New("创建内容文件失败")
		return nil, z.Err
	}
	defer distContentFile.Close()

	distContentFile.WriteString(z.zipInfo.Cmd.String())
	distContentFile.WriteString(";")
	distContentFile.WriteString(base64.StdEncoding.EncodeToString(appInfoBytes))
	distContentFile.WriteString(";")
	distContentFile.WriteString(base64.StdEncoding.EncodeToString(appVersionBytes))
	distContentFile.WriteString(";")
	if z.jdkContentPath != "" {
		err = encryptFrameFile2File("jdk", z.jdkContentPath, distContentFile, z.pubKey)
		if err != nil {
			z.Err = err
			return nil, err
		}
	} else {
		if z.jarPass != nil {
			marshal, err = json.Marshal(z.jarPass)
			if err != nil {
				return nil, errors.New("转换启动密码失败")
			}
			distContentFile.WriteString(base64.StdEncoding.EncodeToString(marshal))
			distContentFile.WriteString(";")
		}

		if z.encJarContentPath != "" {
			err = encryptFrameFile2File("jar", z.encJarContentPath, distContentFile, z.pubKey)
			if err != nil {
				z.Err = err
				return nil, err
			}
		}

		if len(z.pluginsContentPath) > 0 {
			for _, content := range z.pluginsContentPath {
				err = encryptFrameFile2File("plugin", content, distContentFile, z.pubKey)
				if err != nil {
					z.Err = err
					return nil, err
				}
			}
		}
	}
	distFilePath := filepath.Join(z.operationDir, "end")
	_ = os.RemoveAll(distFilePath)
	distFile, err := os.Create(distFilePath)
	if err != nil {
		z.Err = errors.New("创建目标文件失败")
		return nil, z.Err
	}
	defer distFile.Close()

	contentDataMd5Sum, err := hash.CalcMd5(distContentFilePath)
	if err != nil {
		z.Err = errors.New("计算数据包MD5摘要失败")
		return nil, z.Err
	}
	contentDataSha1Sum, err := hash.CalcSha1(distContentFilePath)
	if err != nil {
		z.Err = errors.New("计算数据包SH1摘要失败")
		return nil, z.Err
	}

	dataKey := gmsm.Sm4RandomKey()
	distContentFile.Seek(0, 0)

	distContentEncFilePath := filepath.Join(z.operationDir, "ct_e")
	_ = os.RemoveAll(distContentFilePath)
	distContentEncFile, err := os.Create(distContentEncFilePath)
	if err != nil {
		z.Err = errors.New("创建数据加密文件失败")
		return nil, z.Err
	}
	defer distContentEncFile.Close()

	if err = gmsm.Sm4Encrypt2File(dataKey, distContentFile, distContentEncFile); err != nil {
		z.Err = errors.New("加密数据失败")
		return nil, z.Err
	}

	sm2EncryptKeyData, err := gmsm.Sm2Encrypt(z.pubKey, dataKey)
	if err != nil {
		return nil, errors.New("加密密钥失败")
	}

	distFile.Write(contentDataMd5Sum[:])
	distFile.Write(contentDataSha1Sum[:])
	distFile.Write(sm2EncryptKeyData)
	distContentEncFile.Seek(0, 0)
	if _, err = io.Copy(distFile, distContentEncFile); err != nil {
		return nil, errors.New("写出文件到数据包失败")
	}

	tmpDir := filepath.Join(z.operationDir, "rs")
	_ = os.RemoveAll(tmpDir)
	if err = os.MkdirAll(tmpDir, 0777); err != nil {
		return nil, errors.New("创建临时目录失败")
	}

	appDataFilePath := filepath.Join(tmpDir, "app.content")

	err = compress.Gzip(distFilePath, appDataFilePath)
	if err != nil {
		return nil, errors.New("压缩程序数据失败")
	}

	certInfo := map[string]string{
		"u":  z.zipInfo.UserCertPem,
		"uk": z.zipInfo.UserKeyPem,
		"r":  z.zipInfo.RootCertPem,
	}

	certInfoJsonStr, _ := json.Marshal(certInfo)
	certInfoTmpFile := filepath.Join(tmpDir, "cert.info")
	err = ioutil.WriteFile(certInfoTmpFile, certInfoJsonStr, 0666)
	if err != nil {
		return nil, errors.New("写出证书信息失败")
	}

	tmpBuffer := bytes.Buffer{}
	tmpBuffer.WriteString(z.zipInfo.LocalServerPort)
	tmpBuffer.WriteRune(';')
	tmpBuffer.WriteString(base64.StdEncoding.EncodeToString(appInfoBytes))
	tmpBuffer.WriteRune(';')
	tmpBuffer.WriteString(base64.StdEncoding.EncodeToString(appVersionBytes))
	appInfoTmpFile := filepath.Join(tmpDir, "app.info")
	err = ioutil.WriteFile(appInfoTmpFile, tmpBuffer.Bytes(), 0666)
	if err != nil {
		return nil, errors.New("写出App信息失败")
	}

	runnerFile := filepath.Join(tmpDir, "appRunner.go")
	if err = ioutil.WriteFile(runnerFile, []byte(startRunnerTemplate), 0666); err != nil {
		return nil, errors.New("写出执行器失败")
	}

	appName := "appRunner"
	if goos == "windos" {
		appName += ".exe"
	}

	environ := os.Environ()
	environ = append(environ, "GOOS="+goos)
	environ = append(environ, "GOARCH="+goarch)
	err = execCmd("go", tmpDir, environ, "mod", "init", "app.runner")
	if err != nil {
		return nil, errors.New("创建编译模块失败")
	}

	err = execCmd("go", tmpDir, environ, "mod", "tidy")
	if err != nil {
		return nil, errors.New("下载执行依赖失败")
	}

	err = execCmd("go", tmpDir, environ, "build", "-ldflags=-s -w", "-o", appName)
	if err != nil {
		return nil, errors.New("执行器编译失败")
	}

	appRunnerFile := filepath.Join(tmpDir, appName)
	file, err := ioutil.ReadFile(appRunnerFile)
	if err != nil {
		return nil, errors.New("读取执行程序失败")
	}

	return file, nil
}

func encryptFrameFile2File(cmd string, srcFilePath string, distFile *os.File, encPublicKeyPem *sm2.PublicKey) error {
	srcFile, err := os.OpenFile(srcFilePath, os.O_RDONLY, 0666)
	if err != nil {
		return errors.New("获取源文件失败")
	}
	defer srcFile.Close()

	dir, err := ioutil.TempDir("", "t*")
	if err != nil {
		return errors.New("创建临时保存目录失败")
	}
	defer os.RemoveAll(dir)

	tmpDist, err := os.Create(filepath.Join(dir, "e"))
	if err != nil {
		return errors.New("创建临时保存文件失败")
	}
	defer tmpDist.Close()

	sm4RandomKey := gmsm.Sm4RandomKey()

	if err = gmsm.Sm4Encrypt2File(sm4RandomKey, srcFile, tmpDist); err != nil {
		return errors.New("生成数据加密文件失败")
	}

	stat, err := tmpDist.Stat()
	if err != nil {
		return errors.New("获取文件状态失败")
	}
	encryptLenStr := strconv.FormatInt(stat.Size(), 10)

	sm2EncryptKeyData, err := gmsm.Sm2Encrypt(encPublicKeyPem, sm4RandomKey)
	if err != nil {
		return errors.New("加密密钥失败")
	}

	distFile.WriteString(cmd)
	distFile.WriteString(";")
	distFile.Write(sm2EncryptKeyData)
	distFile.WriteString(encryptLenStr)
	distFile.WriteString(";")

	tmpDist.Seek(0, 0)
	_, err = io.Copy(distFile, tmpDist)
	if err != nil {
		return errors.New("生成数据包失败")
	}
	return nil
}

func execCmd(cmd string, dir string, env []string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Dir = dir
	command.Env = env
	//output, err := command.CombinedOutput()
	//fmt.Println(string(output))
	return command.Run()
	//return err
}
