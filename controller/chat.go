package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"xunfei/common"
	"xunfei/repository"
	"xunfei/response"
	"xunfei/vo"
)

type IChat interface {
	XunFeiChat(c *gin.Context)
	XunFeiSpark(c *gin.Context)
}
type Chat struct {
	xfRepository repository.IXunFeiRepository
}

func (chat *Chat) XunFeiChat(c *gin.Context) {
	var req vo.XunFeiRequest
	// 请求json绑定
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}
	chat.xfRepository.Chat(c, &req)
}
func (chat *Chat) XunFeiSpark(c *gin.Context) {
	var req vo.XunFeiSparkRequest
	// 请求json绑定
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	// 参数校验
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}
	chat.xfRepository.Spark(c, &req)
}

func NewChatController() IChat {
	return &Chat{
		xfRepository: repository.NewXunFeiRepository(),
	}
}
