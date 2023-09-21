package common

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"regexp"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func GetUserIdByRandom() string {
	userId := uuid.Must(uuid.NewV4(), nil).String()
	h := md5.New()
	h.Write([]byte(userId + GetRandomId(10)))
	return hex.EncodeToString(h.Sum(nil))[0:32]
}

func GetRandomId(length int) (orderId string) {
	// 定义随机字符集
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())
	// 生成随机ID
	id := make([]byte, length)
	for i := 0; i < length; i++ {
		id[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(id)
}

func RandPass(lenNum int) string {
	var chars = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	str := strings.Builder{}
	length := len(chars)
	rand.Seed(time.Now().UnixNano()) //重新播种，否则值不会变
	for i := 0; i < lenNum; i++ {
		str.WriteString(chars[rand.Intn(length)])

	}
	return str.String()
}

func RegexCode(text string) string {
	// 定义正则表达式模式
	startTagPattern := `\[code\]`
	endTagPattern := `\[/code\]`

	// 编译正则表达式
	startTagRegex := regexp.MustCompile(startTagPattern)
	endTagRegex := regexp.MustCompile(endTagPattern)

	// 使用正则表达式进行替换
	modifiedText := startTagRegex.ReplaceAllString(text, "```")
	modifiedText = endTagRegex.ReplaceAllString(modifiedText, "```")
	// 打印替换后的文本
	return modifiedText
}
