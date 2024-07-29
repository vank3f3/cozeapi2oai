package model

type Chunk struct {
	Event          string       `json:"event"`
	Message        ChunkMessage `json:"message"`
	IsFinish       bool         `json:"is_finish"`
	Index          int          `json:"index"`
	ConversationID string       `json:"conversation_id"`
	SeqID          int          `json:"seq_id"`
}

type ChunkMessage struct {
	Role        string `json:"role"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}
