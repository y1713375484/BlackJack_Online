package main

import (
	"balckJack/router"
)

func main() {

	r := router.InitRouter()

	r.Run(":7272") // 监听并在 0.0.0.0:8080 上启动服务
}
