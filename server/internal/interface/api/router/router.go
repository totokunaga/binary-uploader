package router

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"github.com/tomoya.tokunaga/server/internal/interface/api/handler"
)

// Router defines the HTTP router for the API
type Router struct {
	engine            *gin.Engine
	fileUploadHandler *handler.FileUploadHandler
	fileListHandler   *handler.FileListHandler
	fileDeleteHandler *handler.FileDeleteHandler
}

// NewRouter creates a new Router instance
func NewRouter(
	fileUploadHandler *handler.FileUploadHandler,
	fileListHandler *handler.FileListHandler,
	fileDeleteHandler *handler.FileDeleteHandler,
) *Router {
	return &Router{
		engine:            gin.Default(),
		fileUploadHandler: fileUploadHandler,
		fileListHandler:   fileListHandler,
		fileDeleteHandler: fileDeleteHandler,
	}
}

// SetupRoutes sets up the routes for the API
func (r *Router) SetupRoutes() {
	root := r.engine.Group("/")
	root.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "OK")
	})

	// API routes
	api := r.engine.Group("/api")

	v1 := api.Group("/v1")
	v1.POST("/upload/init/:file_name", r.fileUploadHandler.ExecuteInit)
	v1.POST("/upload/:upload_id/:chunk_id", r.fileUploadHandler.Execute)
	v1.DELETE("/:file_name", r.fileDeleteHandler.Execute)
	v1.GET("/files", r.fileListHandler.Execute)

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
