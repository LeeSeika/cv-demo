package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Start() error {
	engine := gin.Default()
	engine.Use(gin.Recovery())

	s.addRoutes(engine)

	server := &http.Server{
		Addr:    s.addr,
		Handler: engine,
	}

	return server.ListenAndServe()
}
