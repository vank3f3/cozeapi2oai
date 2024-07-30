package model

// ChatCompletionRequestBody API 请求体
type ChatCompletionRequestBody struct {
	Messages []ChatCompletionRequestBodyMessage `json:"messages"`
	Model    string                             `json:"model"`
	User     string                             `json:"user,omitempty"`
	Stream   bool                               `json:"stream,omitempty"`
}

type ChatCompletionRequestBodyMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CozeAPIRequestBody Coze API 请求体
type CozeAPIRequestBody struct {
	Query          string        `json:"query"`
	Stream         bool          `json:"stream"`
	ConversationID string        `json:"conversation_id"`
	User           string        `json:"user"`
	BotId          string        `json:"bot_id"`
	ChatHistory    []ChatHistory `json:"chat_history"`
}

type ChatHistory struct {
	Role        string `json:"role"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}
