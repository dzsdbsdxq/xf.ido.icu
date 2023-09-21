package repository

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/url"
	"strings"
	"time"
	"xunfei/common"
	"xunfei/config"
	"xunfei/response"
	"xunfei/vo"
)

type IXunFeiRepository interface {
	Chat(c *gin.Context, req *vo.XunFeiRequest)
	Spark(c *gin.Context, req *vo.XunFeiSparkRequest)
	// 生成参数
	genParams(appId string, domain string, temperature float64, topK int64, maxTokens int64, messages []vo.Message) map[string]interface{}
	//创建鉴权url  apikey 即 hmac username
	assembleAuthUrl(hostUrl string, apiKey, apiSecret string) string
	hmacWithShaToBase64(algorithm, data, key string) string
}

type XunFeiRepository struct{}

func (xf *XunFeiRepository) Chat(c *gin.Context, req *vo.XunFeiRequest) {
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	hostUrl := config.Conf.XunFei.HostUrlV1
	if req.Domain == "generalv2" {
		hostUrl = config.Conf.XunFei.HostUrlV2
	}
	common.Log.Info("Chat:[用户最后一次发送消息]：", req.Question[len(req.Question)-1].Content)
	//握手并建立websocket 连接
	conn, resp, err := d.Dial(xf.assembleAuthUrl(hostUrl, config.ApiKey, config.ApiSecret), nil)
	if err != nil {
		response.Fail(c, err, "fail")
		return
	} else if resp.StatusCode != 101 {
		response.Fail(c, err, "fail")
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	//发送消息
	go func() {
		data := xf.genParams(config.AppId, req.Domain, req.Temperature, int64(req.TopK), int64(req.MaxTokens), req.Question)
		_ = conn.WriteJSON(data)
	}()

	done := make(chan *response.DoneMessage, 100)
	xfError := make(chan string, 10)
	var answer = ""
	go func(conn *websocket.Conn) {
		defer func() {
			_ = conn.Close()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				//记录错误日志
				common.Log.Errorf("读取消息错误,错误原因：%s\n", err.Error())
				//xfError <- err.Error()
				msg := &response.DoneMessage{
					Sid:      common.RandPass(16),
					Role:     "assistant",
					Text:     err.Error(),
					Complete: true,
					Usage:    nil,
				}
				done <- msg
				return
			}
			var data map[string]interface{}
			err1 := jsoniter.Unmarshal(message, &data)
			if err1 != nil {
				//记录错误日志
				common.Log.Errorf("JSON解析消息错误,错误原因：%s\n", err.Error())
				//xfError <- fmt.Sprintf("Error parsing JSON:%s", err)
				msg := &response.DoneMessage{
					Sid:      common.RandPass(16),
					Role:     "assistant",
					Text:     fmt.Sprintf("Error parsing JSON:%s", err),
					Complete: true,
					Usage:    nil,
				}
				done <- msg
				return
			}
			header := data["header"].(map[string]interface{})
			code := header["code"].(float64)
			if code != 0 {
				//xfError <- header["message"].(string)
				//记录错误日志
				common.Log.Errorf("讯飞返回消息错误,错误码：%f,错误原因：%s\n", code, header["message"].(string))
				msg := &response.DoneMessage{
					Sid:      common.RandPass(16),
					Role:     "assistant",
					Text:     header["message"].(string),
					Complete: true,
					Usage:    nil,
				}
				done <- msg
				return
			}
			//解析数据
			payload := data["payload"].(map[string]interface{})
			choices := payload["choices"].(map[string]interface{})

			text := choices["text"].([]interface{})
			role := text[0].(map[string]interface{})["role"].(string)
			answer += text[0].(map[string]interface{})["content"].(string)
			usage := make(map[string]interface{}, 4)
			if payload["usage"] != nil {
				usage = payload["usage"].(map[string]interface{})["text"].(map[string]interface{})
			}
			//common.Log.Info("[消息]：", choices["status"], header["message"].(string), text[0].(map[string]interface{})["content"].(string))
			///fmt.Println("读取消息：", req.Stream, code, choices["status"], header["message"].(string), text[0].(map[string]interface{})["content"].(string))
			if !req.Stream {
				if choices["status"].(float64) == 2 {
					msg := &response.DoneMessage{
						Sid:      header["sid"].(string),
						Role:     role,
						Text:     common.RegexCode(answer),
						Complete: true,
						Usage:    usage,
					}
					done <- msg
					return
				}
			} else {
				msg := &response.DoneMessage{
					Sid:      header["sid"].(string),
					Role:     role,
					Text:     common.RegexCode(text[0].(map[string]interface{})["content"].(string)),
					Complete: false,
					Usage:    usage,
				}
				if choices["status"].(float64) == 2 {
					msg.Complete = true
					done <- msg
					return
				} else {
					done <- msg
				}

			}
		}

	}(conn)
	if req.Stream {
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-done; ok {
				c.SSEvent("message", msg)
				if msg.Complete {
					c.SSEvent("stop", "finish")
				}
				return !msg.Complete
			}
			return false
		})
	} else {
		for {
			select {
			case sts := <-done:
				if sts.Text != "" {
					response.Success(c, sts, "success")
					return
				} else {
					return
				}

			case err := <-xfError:
				response.Fail(c, err, "fail")
				return
			}
		}
	}

}
func (xf *XunFeiRepository) Spark(c *gin.Context, req *vo.XunFeiSparkRequest) {
	d := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	hostUrl := config.Conf.XunFeiSpark.HostUrlSpark
	common.Log.Info("Spark:[用户最后一次发送消息]：", req.Question[len(req.Question)-1].Content)
	fmt.Println(hostUrl + req.AssistantId)
	//握手并建立websocket 连接
	conn, resp, err := d.Dial(xf.assembleAuthUrl(hostUrl+req.AssistantId, config.ApiKey, config.ApiSecret), nil)
	if err != nil {
		response.Fail(c, err, "fail")
		return
	} else if resp.StatusCode != 101 {
		response.Fail(c, err, "fail")
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	//发送消息
	go func() {
		data := xf.genParams(config.AppId, req.Domain, req.Temperature, int64(req.TopK), int64(req.MaxTokens), req.Question)

		_ = conn.WriteJSON(data)
	}()

	sparkDone := make(chan *response.DoneMessage, 100)
	var answer = ""
	go func(conn *websocket.Conn) {
		defer func() {
			_ = conn.Close()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				//记录错误日志
				common.Log.Errorf("读取消息错误,错误原因：%s\n", err.Error())
				msg := &response.DoneMessage{
					Sid:      common.RandPass(16),
					Role:     "assistant",
					Text:     err.Error(),
					Complete: true,
					Usage:    nil,
				}
				sparkDone <- msg
				return
			}
			var data map[string]interface{}
			err1 := jsoniter.Unmarshal(message, &data)
			if err1 != nil {
				//记录错误日志
				common.Log.Errorf("JSON解析消息错误,错误原因：%s\n", err.Error())
				msg := &response.DoneMessage{
					Sid:      common.RandPass(16),
					Role:     "assistant",
					Text:     fmt.Sprintf("Error parsing JSON:%s", err),
					Complete: true,
					Usage:    nil,
				}
				sparkDone <- msg
				return
			}
			fmt.Println(data)
			header := data["header"].(map[string]interface{})
			code := header["code"].(float64)
			if code != 0 {
				//记录错误日志
				common.Log.Errorf("讯飞返回消息错误,错误码：%f,错误原因：%s\n", code, header["message"].(string))
				msg := &response.DoneMessage{
					Sid:      common.RandPass(16),
					Role:     "assistant",
					Text:     header["message"].(string),
					Complete: true,
					Usage:    nil,
				}
				sparkDone <- msg
				return
			}
			//解析数据
			payload := data["payload"].(map[string]interface{})
			choices := payload["choices"].(map[string]interface{})

			text := choices["text"].([]interface{})
			role := text[0].(map[string]interface{})["role"].(string)
			answer += text[0].(map[string]interface{})["content"].(string)
			usage := make(map[string]interface{}, 4)
			if payload["usage"] != nil {
				usage = payload["usage"].(map[string]interface{})["text"].(map[string]interface{})
			}
			//common.Log.Info("[消息]：", choices["status"], header["message"].(string), text[0].(map[string]interface{})["content"].(string))
			//fmt.Println("读取消息：", req.Stream, code, choices["status"], header["message"].(string), text[0].(map[string]interface{})["content"].(string))
			if !req.Stream {
				if choices["status"].(float64) == 2 {
					msg := &response.DoneMessage{
						Sid:      header["sid"].(string),
						Role:     role,
						Text:     common.RegexCode(answer),
						Complete: true,
						Usage:    usage,
					}
					sparkDone <- msg
					return
				}
			} else {
				msg := &response.DoneMessage{
					Sid:      header["sid"].(string),
					Role:     role,
					Text:     common.RegexCode(text[0].(map[string]interface{})["content"].(string)),
					Complete: false,
					Usage:    usage,
				}
				if choices["status"].(float64) == 2 {
					msg.Complete = true
					sparkDone <- msg
					return
				} else {
					sparkDone <- msg
				}

			}
		}

	}(conn)
	if req.Stream {
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-sparkDone; ok {
				c.SSEvent("message", msg)
				if msg.Complete {
					c.SSEvent("stop", "finish")
				}
				return !msg.Complete
			}
			return false
		})
	} else {
		for {
			select {
			case sts := <-sparkDone:
				if sts.Text != "" {
					response.Success(c, sts, "success")
					return
				} else {
					return
				}
			}
		}
	}

}

func (xf *XunFeiRepository) genParams(appId string, domain string, temperature float64, topK int64, maxTokens int64, messages []vo.Message) map[string]interface{} {

	data := map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
		"header": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
			"app_id": appId, // 根据实际情况修改返回的数据结构和字段名
		},
		"parameter": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
			"chat": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
				"domain":      domain,      // 根据实际情况修改返回的数据结构和字段名
				"temperature": temperature, // 根据实际情况修改返回的数据结构和字段名
				"top_k":       topK,        // 根据实际情况修改返回的数据结构和字段名
				"max_tokens":  maxTokens,   // 根据实际情况修改返回的数据结构和字段名
				"auditing":    "default",   // 根据实际情况修改返回的数据结构和字段名
			},
		},
		"payload": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
			"message": map[string]interface{}{ // 根据实际情况修改返回的数据结构和字段名
				"text": messages, // 根据实际情况修改返回的数据结构和字段名
			},
		},
	}
	return data // 根据实际情况修改返回的数据结构和字段名
}

// 创建鉴权url  apikey 即 hmac username
func (xf *XunFeiRepository) assembleAuthUrl(hostUrl string, apiKey, apiSecret string) string {
	ul, err := url.Parse(hostUrl)
	if err != nil {
		fmt.Println(err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	// fmt.Println(sgin)
	//签名结果
	sha := xf.hmacWithShaToBase64("hmac-sha256", sgin, apiSecret)
	// fmt.Println(sha)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	return hostUrl + "?" + v.Encode()
}
func (xf *XunFeiRepository) hmacWithShaToBase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

func NewXunFeiRepository() IXunFeiRepository {
	return &XunFeiRepository{}
}
