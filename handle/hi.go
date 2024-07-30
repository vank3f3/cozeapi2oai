package handle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Hi 接口保活
func Hi(c *gin.Context, model string) {
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
		"choices": []interface{}{map[string]interface{}{"index": 0, "message": map[string]interface{}{"role": "assistant", "content": "Hi !"}, "finish_reason": "stop"}},
		"usage":   usageData,
	})
	return
}
