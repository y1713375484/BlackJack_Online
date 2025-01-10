package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type IndexController struct{}

func (this *IndexController) Index(c *gin.Context) {
	websocket := os.Getenv("WEBSOCKET")
	c.HTML(http.StatusOK, "index/index.html", gin.H{
		"websocket": websocket,
	})
}
