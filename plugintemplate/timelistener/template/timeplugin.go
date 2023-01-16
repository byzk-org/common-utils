package main

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
}
