package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"goto-bangumi/internal/core"
)

func main() {
	logDir := "./data"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt,syscall.SIGTERM)

	go func() {
		<-c
		cancel()
	}()
	core.InitProgram(ctx)
	program := core.Program{}
	program.Start(ctx)
	<-ctx.Done()
	// 启动 API 服务器（阻塞）
	// server := api.NewServer()
	// // 或者指定端口: server := api.NewServerWithPort(8080)
	// if err := server.Run(); err != nil {
	// 	panic(err)
	// }
}
