package handle

import (
	models "cozeapi2oai/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

func NoStream(c *gin.Context, body io.Reader, model string) {
	var responseData models.Conversation

	if err := json.NewDecoder(body).Decode(&responseData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing JSON response."})
		return
	}

	if responseData.Code != 0 || responseData.Msg != "success" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected response from Coze API.", "message": responseData.Msg})
		return
	}

	answerMessage := findAnswerMessage(responseData.Messages)

	if answerMessage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No answer message found."})
		return
	}

	result := answerMessage.Content
	usageData := map[string]int{
		"prompt_tokens":     100,
		"completion_tokens": 10,
		"total_tokens":      110,
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []interface{}{map[string]interface{}{"index": 0, "message": map[string]interface{}{"role": "assistant", "content": result}, "finish_reason": "stop"}},
		"usage":   usageData,
	})

	return
}

func findAnswerMessage(messages []models.Message) (answer *models.Message) {
	for _, m := range messages {
		if m.Role == "assistant" && m.Type == "answer" {
			return &models.Message{
				Role:        m.Role,
				Type:        m.Type,
				Content:     m.Content,
				ContentType: m.ContentType,
			}
		}
	}
	return nil
}
