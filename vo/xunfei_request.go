package vo

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type XunFeiRequest struct {
	//必传，讯飞版本，取值为[general,generalv2]，指定访问的领域,general指向V1.5版本 generalv2指向V2版本。注意：不同的取值对应的url也不一样！
	Domain string `json:"domain" form:"domain"`
	//否，取值为[0,1],默认为0.5,核采样阈值。用于决定结果随机性，取值越高随机性越强即相同的问题得到的不同答案的可能性越高
	Temperature float64 `json:"temperature" form:"temperature"`
	//否，取值为[1,4096]，默认为2048，模型回答的tokens的最大长度
	MaxTokens int `json:"max_tokens" form:"max_tokens"`
	//否，取值为[1，6],默认为4，从k个候选中随机选择⼀个（⾮等概率）
	TopK int `json:"top_k" form:"top_k"`
	//否,需要保障用户下的唯一性,用于关联用户会话
	ChatId string `json:"chat_id" form:"chat_id"`
	//否，是否流式响应，默认否
	Stream bool `json:"stream" form:"stream"`
	//问题列表
	Question []Message `json:"question" form:"question" validate:"required"`
}
type XunFeiSparkRequest struct {
	//必传，讯飞版本，取值为[general,generalv2]，指定访问的领域,general指向V1.5版本 generalv2指向V2版本。注意：不同的取值对应的url也不一样！
	Domain string `json:"domain" form:"domain"`
	//assistant_id
	AssistantId string `json:"assistant_id" form:"assistant_id"`
	//否，取值为[0,1],默认为0.5,核采样阈值。用于决定结果随机性，取值越高随机性越强即相同的问题得到的不同答案的可能性越高
	Temperature float64 `json:"temperature" form:"temperature"`
	//否，取值为[1,4096]，默认为2048，模型回答的tokens的最大长度
	MaxTokens int `json:"max_tokens" form:"max_tokens"`
	//否，取值为[1，6],默认为4，从k个候选中随机选择⼀个（⾮等概率）
	TopK int `json:"top_k" form:"top_k"`
	//否,需要保障用户下的唯一性,用于关联用户会话
	ChatId string `json:"chat_id" form:"chat_id"`
	//否，是否流式响应，默认否
	Stream bool `json:"stream" form:"stream"`
	//问题列表
	Question []Message `json:"question" form:"question" validate:"required"`
}
