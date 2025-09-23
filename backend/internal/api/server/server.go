package server

import (
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
)

func New(addr string, router *ginext.Engine) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
