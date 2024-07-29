package handle

import (
	"bufio"
	"bytes"
	models "cozeapi2oai/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var cozeApiBase = os.Getenv("COZE_API_BASE")
var botConfig = make(map[string]string)
var defaultBotID = os.Getenv("BOT_ID")

func ChatCompletions(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "errmsg": "Unauthorized."})
		return
	}

	token := authHeader[len("Bearer "):]
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "errmsg": "Unauthorized."})
		return
	}

	var requestBody struct {
		Messages []map[string]string `json:"messages"`
		Model    string              `json:"model"`
		User     string              `json:"user,omitempty"`
		Stream   bool                `json:"stream,omitempty"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	log.Println("## Messages:", requestBody.Messages)

	user := requestBody.User
	if user == "" {
		user = "apiuser"
	}

	chatHistory := make([]map[string]interface{}, 0, len(requestBody.Messages)-1)
	for _, message := range requestBody.Messages[:len(requestBody.Messages)-1] {
		chatHistory = append(chatHistory, map[string]interface{}{
			"role":         message["role"],
			"content":      message["content"],
			"content_type": "text",
		})
	}

	lastMessage := requestBody.Messages[len(requestBody.Messages)-1]
	queryString := lastMessage["content"]
	stream := requestBody.Stream
	botID := defaultBotID

	if model, exists := botConfig[requestBody.Model]; exists {
		botID = model
	}

	apiRequestBody := map[string]interface{}{
		"query":           queryString,
		"stream":          stream,
		"conversation_id": "",
		"user":            user,
		"bot_id":          botID,
		"chat_history":    chatHistory,
	}

	cozeApiURL := fmt.Sprintf("https://%s/open_api/v2/chat", cozeApiBase)

	// 将map编码为JSON
	jsonData, err := json.Marshal(apiRequestBody)
	if err != nil {
		log.Fatalf("Error occurred in JSON marshal. Err: %s", err)
	}

	req, err := http.NewRequest("POST", cozeApiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error occurred creating request. Err: %s", err)
	}

	// 添加请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Connection", "Keep-alive")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint. Err: %s", err)
	}

	defer resp.Body.Close()

	if stream {
		streamResp(c, resp.Body, requestBody.Model)
	} else {
		noStreamResp(c, resp.Body, requestBody.Model)
	}
}

// 处理流式响应的逻辑
func streamResp(c *gin.Context, body io.Reader, model string) {
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
		lines := splitLines(buffer)

		for _, line := range lines[:len(lines)-1] {
			line = strings.TrimSpace(line)

			if !startsWithData(line) {
				continue
			}

			line = trimPrefix(line, "data:")
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
						sendChunk(c, model, chunkContent)
					}
				}
			case "done":
				sendDone(c, model)
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

func sendDone(c *gin.Context, model string) {
	chunkId := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	chunkCreated := time.Now().Unix()
	finishReason := "stop"

	respData := &models.RespData{
		ID:      chunkId,
		Object:  "chat.completion.chunk",
		Created: chunkCreated,
		Model:   model,
		Choices: []models.RespDataChoice{
			models.RespDataChoice{
				Index:        0,
				Delta:        models.RespDataChoiceDelta{},
				FinishReason: &finishReason,
			},
		},
	}

	c.SSEvent("done", respData)
	c.String(http.StatusOK, "data: [DONE]\n\n")
	c.Writer.Flush()
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

func splitLines(buffer string) []string {
	return strings.Split(buffer, "\n")
}

func startsWithData(line string) bool {
	return strings.HasPrefix(line, "data:")
}

func trimPrefix(line, prefix string) string {
	return strings.TrimPrefix(line, prefix)
}

func sendChunk(c *gin.Context, model, content string) {
	chunkId := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	chunkCreated := time.Now().Unix()

	respData := &models.RespData{
		ID:      chunkId,
		Object:  "chat.completion.chunk",
		Created: chunkCreated,
		Model:   model,
		Choices: []models.RespDataChoice{
			models.RespDataChoice{
				Index: 0,
				Delta: models.RespDataChoiceDelta{
					Content: content,
				},
				FinishReason: nil,
			},
		},
	}

	c.SSEvent("message", respData)
}

func noStreamResp(c *gin.Context, body io.Reader, model string) {
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
