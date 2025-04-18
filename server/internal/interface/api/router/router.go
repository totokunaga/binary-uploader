package router

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"github.com/tomoya.tokunaga/server/internal/interface/api/handler"
)

// Router defines the HTTP router for the API
type Router struct {
	engine      *gin.Engine
	fileHandler *handler.FileHandler
}

// NewRouter creates a new Router instance
func NewRouter(fileHandler *handler.FileHandler) *Router {
	return &Router{
		engine:      gin.Default(),
		fileHandler: fileHandler,
	}
}

// SetupRoutes sets up the routes for the API
func (r *Router) SetupRoutes() {
	api := r.engine.Group("/api/v1")

	// File upload routes
	api.POST("/upload/init/:file_name", r.fileHandler.InitUpload)
	api.POST("/upload/:upload_id/:chunk_id", r.fileHandler.UploadChunk)

	// File management routes
	api.DELETE("/:file_name", r.fileHandler.DeleteFile)
	api.GET("/files", r.fileHandler.ListFiles)

	// Debug routes for profiling
	debug := r.engine.Group("/debug")
	debug.GET("/pprof/", gin.WrapF(pprof.Index))
	debug.GET("/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	debug.GET("/pprof/profile", gin.WrapF(pprof.Profile))
	debug.GET("/pprof/symbol", gin.WrapF(pprof.Symbol))
	debug.GET("/pprof/trace", gin.WrapF(pprof.Trace))
}

// Engine returns the Gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
