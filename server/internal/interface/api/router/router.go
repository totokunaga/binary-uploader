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
	fileGetHandler    *handler.FileGetHandler
	fileDeleteHandler *handler.FileDeleteHandler
}

// NewRouter creates a new Router instance
func NewRouter(
	fileUploadHandler *handler.FileUploadHandler,
	fileGetHandler *handler.FileGetHandler,
	fileDeleteHandler *handler.FileDeleteHandler,
) *Router {
	return &Router{
		engine:            gin.Default(),
		fileUploadHandler: fileUploadHandler,
		fileGetHandler:    fileGetHandler,
		fileDeleteHandler: fileDeleteHandler,
	}
}

// SetupRoutes sets up the routes for the API
func (r *Router) SetupRoutes() {
	// Health check route for the load balancer or monitoring tools
	root := r.engine.Group("/")
	root.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "OK")
	})

	// API routes. A new router group is created for each version of the API for the backward compatibility.
	api := r.engine.Group("/api")

	v1 := api.Group("/v1")
	v1.POST("/files/upload/init/:file_name", r.fileUploadHandler.ExecuteInit)
	v1.POST("/files/upload/:file_id/:chunk_number", r.fileUploadHandler.Execute)
	v1.DELETE("/files/:file_name", r.fileDeleteHandler.Execute)
	v1.GET("/files", r.fileGetHandler.Execute)
	v1.GET("/files/:file_name", r.fileGetHandler.ExecuteGetStats)

	// Debug routes for profiling
	debug := r.engine.Group("/debug")
	debug.GET("/pprof/", gin.WrapF(pprof.Index))
	debug.GET("/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	debug.GET("/pprof/profile", gin.WrapF(pprof.Profile))
	debug.GET("/pprof/symbol", gin.WrapF(pprof.Symbol))
	debug.GET("/pprof/trace", gin.WrapF(pprof.Trace))
}

// Engine returns the Gin engine. Used for dependency injection.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
