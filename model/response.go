package model

type RespData struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []RespDataChoice `json:"choices"`
}

type RespDataChoice struct {
	Index        int                 `json:"index"`
	Delta        RespDataChoiceDelta `json:"delta"`
	FinishReason *string             `json:"finish_reason"`
}

type RespDataChoiceDelta struct {
	Content string `json:"content,omitempty"`
}
