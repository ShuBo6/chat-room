package router

import (
	"github.com/gin-gonic/gin"
	"ws-chat/pkg/controller"
)

var Router *gin.Engine

func InitRouter() {
	Router = gin.New()
	//Router.Use(middleware.Recovery(), middleware.LogMiddleware())
	Router.GET("/websocket", controller.Websocket)
	Router.GET("/", controller.HttpTest)

	Router.Run(":60000")

}
