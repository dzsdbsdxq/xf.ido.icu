package utils

import (
	"encoding/base64"
)

// EncodeStr2Base64 加密base64字符串
func EncodeStr2Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// DecodeStrFromBase64 解密base64字符串
func DecodeStrFromBase64(str string) ([]byte, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(str)
	return decodeBytes, err
}
