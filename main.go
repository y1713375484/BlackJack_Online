package main

import (
	"balckJack/router"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	r := router.InitRouter()
	r.Run(":" + os.Getenv("PORT") + "") // 监听并在 0.0.0.0:8080 上启动服务
}
