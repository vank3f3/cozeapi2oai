package model

type Conversation struct {
	Messages       []Message `json:"messages"`
	ConversationID string    `json:"conversation_id"`
	Code           int       `json:"code"`
	Msg            string    `json:"msg"`
}

type Message struct {
	Role        string `json:"role"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}
