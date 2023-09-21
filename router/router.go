package router

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"xunfei/common"
	"xunfei/config"
	"xunfei/controller"
	"xunfei/response"
)

func InitRouter() *gin.Engine {
	gin.SetMode(config.Conf.System.Mode)
	r := gin.Default()
	r.Use(httpCors())

	initWebRouters(r)

	chatController := controller.NewChatController()
	r.Use(auth())
	r.POST("/xf/chat", chatController.XunFeiChat)
	r.POST("/xf/spark", chatController.XunFeiSpark)
	common.Log.Info("初始化路由完成！")
	return r
}

func initWebRouters(r *gin.Engine) *gin.Engine {
	r.Static("assets", "web/assets")
	r.Static("favicon.ico", "web/favicon.ico")
	r.LoadHTMLGlob("./web/*.html")
	r.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", gin.H{})
	})
	return r
}

func auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorization := context.GetHeader("Authorization")
		fmt.Println("Authorization:", authorization)
		if authorization == "" {
			response.Fail(context, nil, "Access Denied")
			context.Abort()
			return
		}
		//Authorization加密规则： //对Authorization进行解密
		plainText, err := func(encodeStr string) ([]byte, error) {
			originData, err := base64.StdEncoding.DecodeString(encodeStr)
			if err != nil {
				return nil, err
			}
			iv := []byte(config.Conf.Encode.AesIv)
			keyByteString := config.Conf.Encode.AesKey
			cipherBlock, err := aes.NewCipher([]byte(keyByteString))
			if err != nil {
				return nil, err
			}
			cipher.NewCBCDecrypter(cipherBlock, iv).CryptBlocks(originData, originData)
			pckS := func(src []byte) []byte {
				length := len(src)
				unPadding := int(src[length-1])
				if length-unPadding < 0 {
					return []byte("")
				}
				return src[:(length - unPadding)]
			}(originData)
			return pckS, nil
		}(authorization)

		if err != nil {
			response.Fail(context, nil, err.Error())
			context.Abort()
			return
		}
		//对plainText按:进行切割
		sk := strings.Split(string(plainText), ":")
		if len(sk) != 3 {
			response.Fail(context, nil, "Authorization Error")
			context.Abort()
			return
		}
		config.AppId = sk[0]
		config.ApiKey = sk[1]
		config.ApiSecret = sk[2]
		common.Log.Infof("AppId:[%s],ApiKey:[%s],ApiSecret:[%s],鉴权成功", sk[0], sk[1], sk[2])
		context.Next()
	}
}

func httpCors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}
