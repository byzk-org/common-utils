package random

import (
	"math/rand"
	"time"
)

func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz~！@#￥%……&*（）——+」|「P:>?/*-+.+*_*+我爱中国^_^"
	//str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []rune(str)
	result := make([]rune, l, l)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result[i] = bytes[r.Intn(len(bytes))]
		//result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
