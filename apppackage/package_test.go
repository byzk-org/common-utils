package apppackage

import (
	"encoding/pem"
	"fmt"
	"github.com/byzk-org/common-utils/plugintemplate"
	"github.com/tjfoc/gmsm/x509"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	userCert = `-----BEGIN CERTIFICATE-----
MIICDDCCAbGgAwIBAgIIFnD/u564J4AwCgYIKoEcz1UBg3UwaDELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaUppbmcxEDAOBgNVBAcTB0hhaURpYW4xDTALBgNVBAoT
BGJ5emsxDTALBgNVBAsTBGJ5emsxFzAVBgNVBAMMDua1i+ivlUNh6K+B5LmmMCAX
DTIxMDMzMDAzMDQwN1oYDzIxMjEwMzMwMDMwNDA3WjBsMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpSmluZzEQMA4GA1UEBxMHSGFpRGlhbjENMAsGA1UEChMEYnl6
azENMAsGA1UECxMEYnl6azEbMBkGA1UEAwwS5rWL6K+V55So5oi36K+B5LmmMFkw
EwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEXIHQIJR1U8QW/c1BRT9zKz6vmtphT01D
UW94PVoREAa4Y2egEnOaxWLvvp55AIXQpa+80RWG1Xat668tDzYiU6M/MD0wDgYD
VR0PAQH/BAQDAgXgMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDBDAMBgNV
HRMBAf8EAjAAMAoGCCqBHM9VAYN1A0kAMEYCIQDK4FdOgzpb8x8AjK711+H+xqSU
EpfOm2aKxjUpnkYZugIhAJQpMUEW4bKvg7B/7fiE7td6zbYbY+CEuqOb3UjML6zg
-----END CERTIFICATE-----`
	userKey = `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgHkXjw7CmDJSyPxRB
SUOaJpDBB0/eJXyCAzjBVHzceoSgCgYIKoEcz1UBgi2hRANCAARcgdAglHVTxBb9
zUFFP3MrPq+a2mFPTUNRb3g9WhEQBrhjZ6ASc5rFYu++nnkAhdClr7zRFYbVdq3r
ry0PNiJT
-----END PRIVATE KEY-----`
	rootCert = `-----BEGIN CERTIFICATE-----
MIICCjCCAbCgAwIBAgIIFnD/u53WVmswCgYIKoEcz1UBg3UwaDELMAkGA1UEBhMC
Q04xEDAOBgNVBAgTB0JlaUppbmcxEDAOBgNVBAcTB0hhaURpYW4xDTALBgNVBAoT
BGJ5emsxDTALBgNVBAsTBGJ5emsxFzAVBgNVBAMMDua1i+ivlUNh6K+B5LmmMCAX
DTIxMDMzMDAzMDQwN1oYDzIxMjEwMzMwMDMwNDA3WjBoMQswCQYDVQQGEwJDTjEQ
MA4GA1UECBMHQmVpSmluZzEQMA4GA1UEBxMHSGFpRGlhbjENMAsGA1UEChMEYnl6
azENMAsGA1UECxMEYnl6azEXMBUGA1UEAwwO5rWL6K+VQ2Hor4HkuaYwWTATBgcq
hkjOPQIBBggqgRzPVQGCLQNCAARzfSOqQsFHm3ZZbiLsLJ3uXBeon9kz6Yai5LPd
9/2Axyj1wcCzHwy0qqS3TM3xGPHG8Q/vTMDGwXfkr2Brh14Ro0IwQDAOBgNVHQ8B
Af8EBAMCAQYwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMA8GA1UdEwEB
/wQFMAMBAf8wCgYIKoEcz1UBg3UDSAAwRQIhAPWVbnWho+tj+4Rwu4pYAeUZm8RC
s+mz+763TbzHFe01AiAY1rHlJhw2K6ZqYak2ksVFFrhwuUSrCcK/46srr07GZw==
-----END CERTIFICATE-----`
	rootKey = `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQg7emVjfncTU3ijXJw
nKPL6jGX+U7nhnxMICUWSQ7TFhmgCgYIKoEcz1UBgi2hRANCAARzfSOqQsFHm3ZZ
biLsLJ3uXBeon9kz6Yai5LPd9/2Axyj1wcCzHwy0qqS3TM3xGPHG8Q/vTMDGwXfk
r2Brh14R
-----END PRIVATE KEY-----`
)

type appAndVersionInfo struct {
	Name         string          `json:"name"`
	Desc         string          `json:"desc"`
	EnvConfig    []*AppEnvConfig `json:"envConfig,omitempty"`
	JdkStartArgs []string        `json:"jdkStartArgs,omitempty"`
}

func (a *appAndVersionInfo) GetName() string {
	return a.Name
}

func (a *appAndVersionInfo) GetDesc() string {
	return a.Desc
}

func (a *appAndVersionInfo) GetEnvConfig() []*AppEnvConfig {
	return a.EnvConfig
}

func (a *appAndVersionInfo) GetJdkStartArgs() []string {
	return a.JdkStartArgs
}

func TestAppPackageZip(t *testing.T) {

	type pluginListener struct {
		host      string
		timeOut   int
		timeSpace int
	}

	defaultExpire := "2021-10-01 00:00:00"
	var testData = []struct {
		jarFile           string
		name              string
		desc              string
		versionName       string
		versionDesc       string
		envConfig         []*AppEnvConfig
		exclude           []string
		include           []string
		expire            string
		os                string
		arch              string
		beforeListener    []*pluginListener
		afterListenerHost []*pluginListener
	}{
		//{
		//	jarFile:     "tmp/view-web-server-manage-V2.0.0.jar",
		//	name:        "dzqz",
		//	desc:        "?????????????????????",
		//	versionName: "1.0.0",
		//	versionDesc: "????????????1.0.0????????????",
		//	expire:      defaultExpire,
		//	//os:          runtime.GOOS,
		//	//arch:        runtime.GOARCH,
		//	os:   "windows",
		//	arch: "amd64",
		//	exclude: []string{
		//		"/static/*/**",
		//		"/views/*/**",
		//	},
		//	envConfig: []*AppEnvConfig{
		//		{
		//			Name:       "WEB_PORT",
		//			DefaultVal: "8188",
		//			Desc:       "????????????",
		//		},
		//		{
		//			Name:       "DB_HOST",
		//			DefaultVal: "192.168.104.125",
		//			Desc:       "???????????????",
		//		},
		//		{
		//			Name:       "DB_PORT",
		//			DefaultVal: "1521",
		//			Desc:       "???????????????",
		//		},
		//		{
		//			Name:       "DB_NAME",
		//			DefaultVal: "orcl",
		//			Desc:       "?????????????????????",
		//		},
		//		{
		//			Name:       "DB_USERNAME",
		//			DefaultVal: "sdzqz",
		//			Desc:       "??????????????????",
		//		},
		//		{
		//			Name:       "DB_PASSWORD",
		//			DefaultVal: "sdzqz",
		//			Desc:       "???????????????",
		//		},
		//		{
		//			Name:       "DB_SOURCES_MINIMUM_IDLE",
		//			DefaultVal: "5",
		//			Desc:       "???????????????????????????",
		//		},
		//		{
		//			Name:       "DB_SOURCES_POOL_SIZE",
		//			DefaultVal: "5",
		//			Desc:       "????????????????????????",
		//		},
		//		{
		//			Name:       "DB_SOURCES_CONNECTION_TIMEOUT",
		//			DefaultVal: "3000",
		//			Desc:       "???????????????????????????",
		//		},
		//		{
		//			Name:       "DB_SOURCES_MAX_LIFETIME",
		//			DefaultVal: "1800000",
		//			Desc:       "???????????????????????????",
		//		},
		//		{
		//			Name:       "UPLOAD_MAX_FILE_SIZE",
		//			DefaultVal: "100MB",
		//			Desc:       "???????????????????????????????????????",
		//		},
		//		{
		//			Name:       "HTTPCLIENT_CONNECTION_TIME_OUT",
		//			DefaultVal: "5000",
		//			Desc:       "httpClient??????????????????",
		//		},
		//		{
		//			Name:       "HTTPCLIENT_CONNECTION_REQUEST_TIME_OUT",
		//			DefaultVal: "5000",
		//			Desc:       "httpClient??????????????????",
		//		},
		//		{
		//			Name:       "HTTPCLIENT_CONNECTION_SOCKET_TIME_OUT",
		//			DefaultVal: "10000",
		//			Desc:       "httpClient Socket????????????",
		//		},
		//		{
		//			Name:       "LOG_LEVEL",
		//			DefaultVal: "INFO",
		//			Desc:       "????????????",
		//		},
		//		{
		//			Name:       "LOG_PATH",
		//			DefaultVal: "/tmp/dzqzLogs",
		//			Desc:       "??????????????????",
		//		},
		//		{
		//			Name:       "LD_LIBRARY_PATH",
		//			DefaultVal: ".",
		//			Desc:       "LD_LIBRARY_PATH",
		//		},
		//	},
		//},
		{
			jarFile:     "demo.jar",
			name:        "test",
			desc:        "????????????",
			versionName: "1.0.0",
			versionDesc: "???????????? 1.0.0",
			expire:      defaultExpire,
			//os:          "windows",
			//arch:        "amd64",
			os:   runtime.GOOS,
			arch: runtime.GOARCH,
			envConfig: []*AppEnvConfig{
				{
					Name:       "test",
					Desc:       "?????????????????????",
					DefaultVal: "abc",
				},
			},
			beforeListener: []*pluginListener{
				{
					host:      ":5000",
					timeOut:   60,
					timeSpace: 5,
				},
			},
		},
		//{
		//	jarFile:     "tmp/xtqm-register.jar",
		//	name:        "xtqm-register",
		//	desc:        "??????????????????",
		//	versionName: "1.0.0",
		//	versionDesc: "?????????????????? 1.0.0",
		//	exclude: []string{
		//		"/static/*/**",
		//		"/static/**/*",
		//	},
		//	expire: defaultExpire,
		//	os:     runtime.GOOS,
		//	arch:   runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/mobileShieldApi.jar",
		//	name:        "mobileShieldApi",
		//	desc:        "??????????????????????????????2",
		//	versionName: "1.0.0",
		//	versionDesc: "??????????????????????????????2 1.0.0",
		//	expire:      defaultExpire,
		//	os:          runtime.GOOS,
		//	arch:        runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/mobileShieldWeb.jar",
		//	name:        "mobileShieldWeb",
		//	desc:        "??????????????????????????????",
		//	versionName: "1.0.0",
		//	versionDesc: "??????????????????????????????",
		//	include: []string{
		//		"/com/sansec/**/*.class",
		//	},
		//	expire: defaultExpire,
		//	os:     runtime.GOOS,
		//	arch:   runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/xtqm-auth.jar",
		//	name:        "xtqm-auth",
		//	desc:        "?????????????????????????????????",
		//	versionName: "1.0.0",
		//	versionDesc: "???????????????????????? 1.0.0",
		//	expire:      defaultExpire,
		//	os:          runtime.GOOS,
		//	arch:        runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/xtqm-gateway.jar",
		//	name:        "xtqm-gateway",
		//	desc:        "????????????API??????",
		//	versionName: "1.0.0",
		//	versionDesc: "????????????API?????? 1.0.0",
		//	expire:      defaultExpire,
		//	os:          runtime.GOOS,
		//	arch:        runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/mobileShieldRS.jar",
		//	name:        "mobileShieldRS",
		//	desc:        "??????????????????????????????1",
		//	versionName: "1.0.0",
		//	versionDesc: "??????????????????????????????1 1.0.0",
		//	expire:      defaultExpire,
		//	os:          runtime.GOOS,
		//	arch:        runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/mobileShieldCM.jar",
		//	name:        "mobileShieldCM",
		//	desc:        "??????????????????????????????",
		//	versionName: "1.0.0",
		//	versionDesc: "??????????????????????????????",
		//	expire:      defaultExpire,
		//	os:          runtime.GOOS,
		//	arch:        runtime.GOARCH,
		//},
		//{
		//	jarFile:     "tmp/mobileShield.jar",
		//	name:        "mobileShield",
		//	desc:        "?????????????????????",
		//	versionName: "1.0.0",
		//	versionDesc: "????????????????????? 1.0.0",
		//	expire:      defaultExpire,
		//	os:          runtime.GOOS,
		//	arch:        runtime.GOARCH,
		//},
	}

	dir, err := ioutil.TempDir("", "pluginSave*")
	if err != nil {
		t.Error("????????????????????????")
		return
	}
	defer os.RemoveAll(dir)
	for _, d := range testData {
		operation := NewZipOperation(&ZipInfo{
			Cmd:             PackTypeInstallApp,
			JarFile:         d.jarFile,
			RootCertPem:     rootCert,
			UserCertPem:     userCert,
			UserKeyPem:      userKey,
			EncryptJarUrl:   "192.168.100.22:8888",
			Exclude:         d.exclude,
			Include:         d.include,
			LocalServerPort: "65529",
			AppInfo: &appAndVersionInfo{
				Name: d.name,
				Desc: d.desc,
			},
			AppVersionInfo: &appAndVersionInfo{
				Name:      d.versionName,
				Desc:      d.versionDesc,
				EnvConfig: d.envConfig,
			},
		})

		block, _ := pem.Decode([]byte(rootKey))
		if block == nil {
			t.Error("??????????????????")
			return
		}
		key, err := x509.ParsePKCS8UnecryptedPrivateKey(block.Bytes)
		if err != nil {
			t.Error("??????????????????")
			return
		}

		//expire, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-06-30 23:59:59", time.Local)
		//expire, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-03-26 18:40:00", time.Local)
		expire, _ := time.ParseInLocation("2006-01-02 15:04:05", d.expire, time.Local)
		pluginOperation := plugintemplate.NewPluginTemplate(key, d.os, d.arch, dir).TimePlugin(expire)
		if len(d.beforeListener) > 0 {
			for _, l := range d.beforeListener {
				if l.timeSpace == 0 {
					l.timeSpace = 5
				}

				if l.timeOut == 0 {
					l.timeOut = 60
				}

				pluginOperation.TelnetBeforePlugin(l.host, l.timeOut, l.timeSpace)
			}
		}

		if len(d.afterListenerHost) > 0 {
			for _, l := range d.afterListenerHost {
				if l.timeSpace == 0 {
					l.timeSpace = 5
				}

				if l.timeOut == 0 {
					l.timeOut = 60
				}

				pluginOperation.TelnetBeforePlugin(l.host, l.timeOut, l.timeSpace)
			}
		}

		result, err := pluginOperation.Result()
		if err != nil {
			t.Error(err.Error())
			return
		}

		for _, r := range result {
			operation.ZipPlugin(r.PluginInfo.String(), r.Content)
		}

		endData, err := operation.
			SetJdkPath("/home/slx/00-Applications/03-java/01java8/jdk1.8.0_181/jre/bin/java").
			SetEncJarPath("/home/slx/work/10-?????????jar??????????????????/common-utils/apppackage/jar-enc.jar").
			ZipApp().
			End(d.os, d.arch)
		if err != nil {
			t.Error(err.Error())
			return
		}

		saveDir := filepath.Join("tmp", "build")
		_ = os.MkdirAll(saveDir, 0777)

		fileSuffixName := ""
		if d.os == "windows" {
			fileSuffixName = ".exe"
		}
		filePath := filepath.Join(saveDir, d.name+"-"+d.versionName+"-"+d.os+"-"+d.arch+fileSuffixName)
		err = ioutil.WriteFile(filePath, endData, 0777)
		if err != nil {
			fmt.Println("???????????? => " + d.name + "-" + d.versionName)
			continue
		}
		fmt.Println("???????????? => " + d.name + "-" + d.versionName)
	}
}

func TestPluginPackageZip(t *testing.T) {

	type pluginListener struct {
		host      string
		timeOut   int
		timeSpace int
	}

	defaultExpire := "2021-10-01 00:00:00"
	var testData = []struct {
		name              string
		desc              string
		versionName       string
		versionDesc       string
		os                string
		arch              string
		beforeListener    []*pluginListener
		afterListenerHost []*pluginListener
		expire            string
	}{
		{
			name:        "xtqm-register",
			desc:        "??????????????????",
			versionName: "1.0.0",
			versionDesc: "?????????????????? 1.0.0",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "mobileShieldApi",
			desc:        "??????????????????????????????2",
			versionName: "1.0.0",
			versionDesc: "??????????????????????????????2 1.0.0",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "mobileShieldWeb",
			desc:        "??????????????????????????????",
			versionName: "1.0.0",
			versionDesc: "??????????????????????????????",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "xtqm-auth",
			desc:        "?????????????????????????????????",
			versionName: "1.0.0",
			versionDesc: "???????????????????????? 1.0.0",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "xtqm-gateway",
			desc:        "????????????API??????",
			versionName: "1.0.0",
			versionDesc: "????????????API?????? 1.0.0",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "mobileShieldRS",
			desc:        "??????????????????????????????1",
			versionName: "1.0.0",
			versionDesc: "??????????????????????????????1 1.0.0",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "mobileShieldCM",
			desc:        "??????????????????????????????",
			versionName: "1.0.0",
			versionDesc: "??????????????????????????????",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
		{
			name:        "mobileShield",
			desc:        "?????????????????????",
			versionName: "1.0.0",
			versionDesc: "????????????????????? 1.0.0",
			expire:      defaultExpire,
			os:          runtime.GOOS,
			arch:        runtime.GOARCH,
			beforeListener: []*pluginListener{
				{
					timeOut:   6000,
					timeSpace: 5,
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "pluginSave*")
	if err != nil {
		t.Error("????????????????????????")
		return
	}
	defer os.RemoveAll(dir)
	for _, d := range testData {

		operation := NewZipOperation(&ZipInfo{
			Cmd:             PackTypeInstallPlugin,
			RootCertPem:     rootCert,
			UserCertPem:     userCert,
			UserKeyPem:      userKey,
			LocalServerPort: "65529",
			AppInfo: &appAndVersionInfo{
				Name: d.name,
				Desc: d.versionName,
			},
			AppVersionInfo: &appAndVersionInfo{
				Name: d.versionName,
				Desc: d.versionDesc,
			},
		})

		block, _ := pem.Decode([]byte(rootKey))
		if block == nil {
			t.Error("??????????????????")
			return
		}
		key, err := x509.ParsePKCS8UnecryptedPrivateKey(block.Bytes)
		if err != nil {
			t.Error("??????????????????")
			return
		}

		pluginTemplate := plugintemplate.NewPluginTemplate(key, d.os, d.arch, dir)
		if len(d.beforeListener) > 0 {
			for _, l := range d.beforeListener {
				if l.timeSpace == 0 {
					l.timeSpace = 5
				}

				if l.timeOut == 0 {
					l.timeOut = 60
				}

				pluginTemplate.TelnetBeforePlugin(l.host, l.timeOut, l.timeSpace)
			}
		}

		if len(d.afterListenerHost) > 0 {
			for _, l := range d.afterListenerHost {
				if l.timeSpace == 0 {
					l.timeSpace = 5
				}

				if l.timeOut == 0 {
					l.timeOut = 60
				}

				pluginTemplate.TelnetBeforePlugin(l.host, l.timeOut, l.timeSpace)
			}
		}

		result, err := pluginTemplate.Result()
		if err != nil {
			t.Error(err.Error())
			return
		}

		for _, r := range result {
			operation.ZipPlugin(r.PluginInfo.String(), r.Content)
		}

		endData, err := operation.End(runtime.GOOS, runtime.GOARCH)
		if err != nil {
			t.Error(err.Error())
			return
		}

		saveDir := filepath.Join("tmp", "build", "plugins")
		_ = os.MkdirAll(saveDir, 0777)
		fileSuffixName := ""
		if d.os == "windows" {
			fileSuffixName = ".exe"
		}

		filePath := filepath.Join(saveDir, d.name+"-"+d.versionName+"-"+d.os+"-plugin-"+d.arch+fileSuffixName)
		err = ioutil.WriteFile(filePath, endData, 0666)
		if err != nil {
			fmt.Println("???????????? => " + d.name + "-" + d.versionName)
			continue
		}
		fmt.Println("???????????? => " + d.name + "-" + d.versionName)
	}

}

func TestJdkPackage(t *testing.T) {
	jdkInfo := &appAndVersionInfo{
		//Name: "jdk1.8-linux-amd64",
		Name: "jdk1.8-jce",
		Desc: "jdk1.8 linux amd64??????",
	}
	operation := NewZipOperation(&ZipInfo{
		Cmd:             PackTypeInstallJdk,
		RootCertPem:     rootCert,
		UserCertPem:     userCert,
		UserKeyPem:      userKey,
		LocalServerPort: "65529",
		AppInfo:         jdkInfo,
		AppVersionInfo:  jdkInfo,
	})
	//endData, err := operation.ZipJdk("jdk1.8.tar.gz").End(runtime.GOOS, runtime.GOARCH)
	endData, err := operation.ZipJdk("tmp/jdk1.8-jce.tar.gz").End(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Error(err.Error())
		return
	}
	ioutil.WriteFile("tmp/jdk-jce-test-linux-amd64", endData, 0777)
}
