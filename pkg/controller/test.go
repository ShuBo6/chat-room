package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func HttpTest(c *gin.Context) {
	message := c.Query("message")
	if message=="" {
		message="ok"
	}
	fmt.Println(message)
	c.JSON(200, message)
}
