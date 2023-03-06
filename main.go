package main

import (
	"net/http"

	"github.com/fxmbx/go-live-chat/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	go ws.Manager.Start()

	r := gin.Default()
	r.LoadHTMLFiles("client.htm")
	r.GET("/client", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "client.htm", nil)
	})
	r.GET("/ws", ws.WsHandler)
	r.GET("/pong", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8282")
}
