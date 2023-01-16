package telnet

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

const templateStr = `package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	timeOut   = {{ .TimeOut }}
	timeSpace = {{ .TimeSpace }}
)

var (
	host      = "{{ .Host }}"
)

func main() {
	envHost := os.Getenv("telnet_host")
	if envHost != "" {
		host = envHost
	}
	jsonMap := map[string]string{
		"type": "{{ .Type }}",
		"desc": "telnet插件, 端口检测插件, 超时时间: {{ .TimeOut }}秒, 检测间隔: {{ .TimeSpace }}",
	}
	startEnv := os.Getenv("__cmd__")
	switch startEnv {
	case "start":
		timer := time.NewTimer(timeOut * time.Second)
		go func() {
			<-timer.C
			timer.Stop()
			fmt.Println("地址 " + host + " 检测超时")
			os.Exit(1)
		}()
		for {
			fmt.Println("正在检测: " + host)
			if TelnetHost(host) {
				fmt.Println(host + " 连接通过, 插件结束")
				os.Exit(0)
			}
			time.Sleep(timeSpace * time.Second)
		}
	case "info":
		marshal, _ := json.Marshal(jsonMap)
		fmt.Println(hex.EncodeToString(marshal) + ";")
	default:
		fmt.Println(jsonMap["desc"])
	}
}

func Telnet(ip string, port int32) bool {
	var (
		err error
	)

	var conn net.Conn
	if conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), 3*time.Second); err != nil {
		return false
	}

	_ = conn.Close()
	return true
}

func TelnetHost(host string) bool {
	var (
		err error
	)

	var conn net.Conn
	if conn, err = net.DialTimeout("tcp", host, 3*time.Second); err != nil {
		return false
	}

	_ = conn.Close()
	return true
}`

func CreateBefore(host string, timeOut, timeSpace int, goos, goarch, saveDir string) (string, error, []interface{}) {
	return create(host, timeOut, timeSpace, "before", goos, goarch, saveDir)
}

func CreateAfter(host string, timeOut, timeSpace int, goos, goarch, saveDir string) (string, error, []interface{}) {
	return create(host, timeOut, timeSpace, "after", goos, goarch, saveDir)
}

func create(host string, timeOut, timeSpace int, pluginType string, goos, goarch, saveDir string) (string, error, []interface{}) {
	tmp, err := template.New("plugin").Parse(templateStr)
	if err != nil {
		return "", errors.New("解析插件模板失败"), nil
	}

	_ = os.MkdirAll(saveDir, 0777)
	pluginPath := filepath.Join(saveDir, "telnetPlugin.go")
	os.RemoveAll(pluginPath)
	pluginFile, err := os.Create(pluginPath)
	if err != nil {
		return "", errors.New("创建插件保存文件失败"), nil
	}
	defer pluginFile.Close()

	if err = tmp.Execute(pluginFile, map[string]interface{}{
		"Host":      host,
		"TimeOut":   timeOut,
		"TimeSpace": timeSpace,
		"Type":      pluginType,
	}); err != nil {
		return "", errors.New("解析模板失败"), nil
	}

	outName := "telnetPlugin" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if goos == "windows" {
		outName += ".exe"
	}

	env := os.Environ()
	env = append(env, "GOOS="+goos)
	env = append(env, "GOARCH=", goarch)

	command := exec.Command("go", "build", "-ldflags=-s -w", "-o", outName, pluginPath)
	command.Dir = saveDir
	command.Env = env
	if err = command.Run(); err != nil {
		return "", errors.New("插件编译失败 => " + err.Error()), nil
	}

	return filepath.Join(saveDir, outName), nil, []interface{}{
		&struct {
			Name       string `json:"name,omitempty"`
			Desc       string `json:"desc,omitempty"`
			DefaultVal string `json:"defaultVal,omitempty"`
		}{
			Name:       "telnet_host",
			Desc:       "telnet检测端口",
			DefaultVal: "",
		},
	}
}
