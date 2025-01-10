package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type IndexController struct{}

func (this *IndexController) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index/index.html", gin.H{})
}
