package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"goto-bangumi/api/middleware"
	"goto-bangumi/api/routes"
)

// DefaultPort 默认端口
const DefaultPort = 7892

// Server API 服务器
type Server struct {
	router *gin.Engine
	port   int
}

// NewServer 创建 API 服务器
func NewServer() *Server {
	return NewServerWithPort(DefaultPort)
}

// NewServerWithPort 创建指定端口的 API 服务器
func NewServerWithPort(port int) *Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	s := &Server{
		router: r,
		port:   port,
	}

	s.registerRoutes()
	return s
}

// registerRoutes 注册所有路由
func (s *Server) registerRoutes() {
	// API v1 路由组
	v1 := s.router.Group("/api/v1")

	// 公开路由（无需认证）
	routes.RegisterAuthRoutes(v1)

	// 需要认证的路由
	authorized := v1.Group("")
	authorized.Use(middleware.JWTAuth())
	{
		routes.RegisterLogRoutes(authorized)
		routes.RegisterProgramRoutes(authorized)
		routes.RegisterConfigRoutes(authorized)
		routes.RegisterBangumiRoutes(authorized)
		routes.RegisterRSSRoutes(authorized)
		routes.RegisterSearchRoutes(authorized)
		routes.RegisterTorrentRoutes(authorized)
	}
}

// Run 启动服务器
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	slog.Info("API 服务器启动", "port", s.port)
	return http.ListenAndServe(addr, s.router)
}

// Router 返回 gin 路由引擎
func (s *Server) Router() *gin.Engine {
	return s.router
}
