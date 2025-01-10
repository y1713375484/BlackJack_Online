package router

import (
	"balckJack/controller"
	"balckJack/static"
	"balckJack/view"
	"balckJack/websocket"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"html/template"
	"net/http"
)

func InitRouter() *gin.Engine {
	router := gin.Default()
	// 创建一个基于 cookie 的存储，设置密钥用于加密会话数据
	store := cookie.NewStore([]byte("secret"))
	// 使用会话中间件
	router.Use(sessions.Sessions("GameSession", store))
	// 使用自定义的自动生成会话 ID 的中间件
	router.Use(SessionMiddleware())
	router.StaticFS("/static", http.FS(static.StaticFile))
	//加载静态页面并打包到执行文件
	parseFS, err := template.ParseFS(view.ViewFile, "*")
	if err != nil {
		fmt.Println(err)
	}
	router.SetHTMLTemplate(parseFS)
	//router.LoadHTMLGlob("view/*")
	webSocket := &websocket.WebSocket{}
	router.GET("/ws", webSocket.Handfunc)
	indexController := &controller.IndexController{}
	router.GET("/", indexController.Index)

	return router
}

// 自动设置session_id
func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("session_id") == nil {
			sessionID := uuid.New().String()
			session.Set("session_id", sessionID)
			session.Save()
		}
		c.Next()
	}
}
