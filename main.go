// main.go
package main

import (
	"cozeapi2oai/handle"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建一个默认的 Gin 路由器
	router := gin.Default()

	// 设置 Cors
	router.Use(handle.Cors)

	// 首页路由
	router.GET("/", handle.Index)

	// 对话接口
	router.POST("/v1/chat/completions", handle.ChatCompletions)

	// 启动HTTP服务器，监听8080端口
	router.Run(":3001")
}
