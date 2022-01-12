package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"ws-chat/pkg/services"
)

func Websocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		msg := fmt.Sprintf("upgrade connect to websocket failed, error:%s", err.Error())
		logrus.Warningf(msg)
		c.AbortWithStatusJSON(http.StatusBadRequest, msg)
		return
	}
	services.GetServer().AddClient(conn)
}