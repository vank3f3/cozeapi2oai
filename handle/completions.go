package handle

import (
	"bytes"
	models "cozeapi2oai/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
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

	requestBody := models.ChatCompletionRequestBody{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	log.Println("## Messages:", requestBody.Messages)

	// 如果是接口保活
	if requestBody.Messages[0].Content == "hi" || requestBody.Messages[0].Content == "ping" {
		Hi(c, requestBody.Model)
		return
	}

	user := requestBody.User
	if user == "" {
		user = "apiuser"
	}

	// 构建 Coze API 请求体
	lastMessage := requestBody.Messages[len(requestBody.Messages)-1]
	queryString := lastMessage.Content
	stream := requestBody.Stream
	botID := defaultBotID
	if model, exists := botConfig[requestBody.Model]; exists {
		botID = model
	}

	cozeAPIRequestBody := &models.CozeAPIRequestBody{
		Query:       queryString,
		Stream:      stream,
		User:        user,
		BotId:       botID,
		ChatHistory: []models.ChatHistory{},
	}
	for _, message := range requestBody.Messages[:len(requestBody.Messages)-1] {
		cozeAPIRequestBody.ChatHistory = append(cozeAPIRequestBody.ChatHistory, models.ChatHistory{
			Role:        message.Role,
			Content:     message.Content,
			ContentType: "text",
		})
	}

	cozeApiURL := fmt.Sprintf("https://%s/open_api/v2/chat", cozeApiBase)
	// 将map编码为JSON
	jsonData, err := json.Marshal(cozeAPIRequestBody)
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
		Stream(c, resp.Body, requestBody.Model)
	} else {
		NoStream(c, resp.Body, requestBody.Model)
	}
}
