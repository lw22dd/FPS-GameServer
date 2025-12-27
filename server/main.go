package main

import (
	"game/app"
	"log"
)

func main() {
	// 创建并启动服务器
	server := app.NewServer()
	if err := server.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
