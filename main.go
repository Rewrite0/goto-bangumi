package main

import (
	"goto-bangumi/api"
	"goto-bangumi/internal/core"
)

func main() {
	core.InitProgram()
	program := core.Program{}
	program.Start()

	// 启动 API 服务器（阻塞）
	server := api.NewServer()
	// 或者指定端口: server := api.NewServerWithPort(8080)
	if err := server.Run(); err != nil {
		panic(err)
	}
}

