package timelistener

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
	"os"
	"time"
)

var licenseTime, _ = time.ParseInLocation("2006-01-02 15:04:05", "{{ . }}", time.Local)

func main() {
	jsonMap := map[string]string{
		"type": "listener",
		"desc": "时间许可插件, 到期时间 {{ . }}",
	}
	startEnv := os.Getenv("__cmd__")
	switch startEnv {
	case "start":
		sub := licenseTime.Sub(time.Now())
		timer := time.NewTimer(sub)
		<-timer.C
		timer.Stop()
		fmt.Println("许可证已过期")
	case "info":
		marshal, _ := json.Marshal(jsonMap)
		fmt.Println(hex.EncodeToString(marshal) + ";")
	default:
		fmt.Println(jsonMap["desc"])
	}
}`

func Create(t time.Time, goos, goarch, saveDir string) (string, error) {
	timeStr := t.Format("2006-01-02 15:04:05")
	tmp, err := template.New("plugin").Parse(templateStr)
	if err != nil {
		return "", errors.New("解析插件模板失败")
	}

	_ = os.MkdirAll(saveDir, 0777)
	pluginPath := filepath.Join(saveDir, "timePlugin.go")
	_ = os.RemoveAll(pluginPath)
	pluginFile, err := os.Create(pluginPath)
	if err != nil {
		return "", errors.New("创建插件保存文件失败")
	}
	defer pluginFile.Close()

	if err = tmp.Execute(pluginFile, timeStr); err != nil {
		return "", errors.New("解析模板失败")
	}

	outName := "timePlugin" + strconv.FormatInt(time.Now().UnixNano(), 10)
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
		return "", errors.New("插件编译失败 => " + err.Error())
	}

	return filepath.Join(saveDir, outName), nil
}
