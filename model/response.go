package model

type StreamRespData struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []StreamRespDataChoice `json:"choices"`
}

type StreamRespDataChoice struct {
	Index        int                       `json:"index"`
	Delta        StreamRespDataChoiceDelta `json:"delta"`
	FinishReason *string                   `json:"finish_reason"`
}

type StreamRespDataChoiceDelta struct {
	Content string `json:"content,omitempty"`
}

// NoStreamRespData 非流 API 相应
type NoStreamRespData struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []NoStreamRespDataChoice `json:"choices"`
	Usage   NoStreamRespDataUsage    `json:"usage"`
}

type NoStreamRespDataUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type NoStreamRespDataChoice struct {
	Index        int           `json:"index"`
	Message      ChoiceMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type ChoiceMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
