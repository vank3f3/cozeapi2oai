package handle

import (
	"bufio"
	models "cozeapi2oai/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

// Stream  处理流式响应的逻辑
func Stream(c *gin.Context, body io.Reader, model string) {
	c.Header("Content-Type", "text/event-stream")

	scanner := bufio.NewScanner(body)
	//scanner.Split(bufio.ScanLines)
	buffer := ""

	for scanner.Scan() {
		// 为空直接进行下一次
		if scanner.Text() == "" {
			continue
		}

		buffer += scanner.Text() + "\n"
		lines := strings.Split(buffer, "\n")

		for _, line := range lines[:len(lines)-1] {
			line = strings.TrimSpace(line)

			if !strings.HasPrefix(line, "data:") {
				continue
			}

			line = strings.TrimPrefix(line, "data:")
			var chunkObj models.Chunk
			if err := json.Unmarshal([]byte(line), &chunkObj); err != nil {
				fmt.Printf("Error parsing chunk: %v\n", err)
				continue
			}

			switch chunkObj.Event {
			case "message":
				if chunkObj.Message.Role == "assistant" && chunkObj.Message.Type == "answer" {
					chunkContent := chunkObj.Message.Content
					if chunkContent != "" {
						sendChunk(c, model, chunkContent, false)
					}
				}
			case "done":
				sendChunk(c, model, "", true)
				return
			case "error":
				errorMsg := fmt.Sprintf("This is an error: %v\n", chunkObj)
				sendError(c, errorMsg)
				return
			}
		}
		buffer = lines[len(lines)-1]
	}
}

func sendChunk(c *gin.Context, model string, content string, isFinish bool) {
	chunkId := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	chunkCreated := time.Now().Unix()

	var finishReason *string = nil
	if isFinish {
		finishReason = new(string)
		*finishReason = "stop"
	}

	// 未终止则赋值Content
	delta := models.StreamRespDataChoiceDelta{}
	if !isFinish {
		delta = models.StreamRespDataChoiceDelta{
			Content: content,
		}
	}

	respData := &models.StreamRespData{
		ID:      chunkId,
		Object:  "chat.completion.chunk",
		Created: chunkCreated,
		Model:   model,
		Choices: []models.StreamRespDataChoice{
			models.StreamRespDataChoice{
				Index:        0,
				Delta:        delta,
				FinishReason: finishReason,
			},
		},
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// 把 respData 转成 Json
	respDataJson, err := json.Marshal(respData)
	if err != nil {
		panic(err)
	}
	// 手动构建 SSE 消息
	sseMessage := fmt.Sprintf("data: %s\n\n", respDataJson)
	// 写入 SSE 消息到响应中
	c.Writer.Write([]byte(sseMessage))

	// 终止
	if isFinish {
		c.String(http.StatusOK, "data: [DONE]\n\n")
		c.Writer.Flush()
		return
	}

}

func sendError(c *gin.Context, errorMsg string) {
	errorData := map[string]interface{}{
		"error": map[string]interface{}{
			"error":   "Unexpected response from Coze API.",
			"message": errorMsg,
		},
	}

	c.SSEvent("error", errorData)
	c.String(http.StatusOK, "data: [DONE]\n\n")
	c.Writer.Flush()
}
