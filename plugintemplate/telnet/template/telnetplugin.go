package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	timeOut   = 60
	timeSpace = 3
)

var (
	host = "127.0.0.1:5000"
)

func main() {
	envHost := os.Getenv("telnet_host")
	if envHost != "" {
		host = envHost
	}
	jsonMap := map[string]string{
		"type": "before",
		"desc": "telnet插件, 检测目标: " + host,
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
				fmt.Println(host + " 连接通过, 插件推出")
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
}
