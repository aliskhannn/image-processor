package router

import (
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/image-processor/internal/api/handlers/image"
	"github.com/aliskhannn/image-processor/internal/middleware"
)

func Setup(h *image.Handler) *ginext.Engine {
	r := ginext.New()

	r.Use(middleware.CORSMiddleware())
	r.Use(ginext.Logger())
	r.Use(ginext.Recovery())

	api := r.Group("/api")

	api.POST("/upload", h.Upload)      // uploading image
	api.GET("/image/:id", h.Get)       // getting image by id
	api.DELETE("/image/:id", h.Delete) // deleting image by id

	return r
}
