package di

import (
	"os"

	"golang.org/x/exp/slog"
	"gorm.io/gorm"

	"github.com/google/wire"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	"github.com/tomoya.tokunaga/server/internal/infra/database"
	"github.com/tomoya.tokunaga/server/internal/interface/api/handler"
	"github.com/tomoya.tokunaga/server/internal/interface/api/router"
	db_repo "github.com/tomoya.tokunaga/server/internal/interface/repository/database"
	fs_repo "github.com/tomoya.tokunaga/server/internal/interface/repository/storage"
	"github.com/tomoya.tokunaga/server/internal/usecase"
)

// ConfigProvider provides the application configuration
func ConfigProvider() *entity.Config {
	return entity.NewConfig()
}

// LoggerProvider provides the application logger
func LoggerProvider() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}

// ----------------------------------------------------------------
// Infrastructure Providers
// ----------------------------------------------------------------
func DBProvider(config *entity.Config, logger *slog.Logger) (*gorm.DB, error) {
	return database.NewDB(config, logger)
}

// ----------------------------------------------------------------
// Repository Providers
// ----------------------------------------------------------------
func FileRepositoryProvider(db *gorm.DB) db_repo.FileRepository {
	return db_repo.NewFileRepository(db)
}

func StorageRepositoryProvider(config *entity.Config, logger *slog.Logger) fs_repo.FileStorageRepository {
	return fs_repo.NewStorageRepository(config, logger)
}

// ----------------------------------------------------------------
// Usecase Providers
// ----------------------------------------------------------------
func FileGetUseCaseProvider(fileRepo db_repo.FileRepository) usecase.FileGetUseCase {
	return usecase.NewFileGetUseCase(fileRepo)
}

func FileUploadUseCaseProvider(fileRepo db_repo.FileRepository, storageRepo fs_repo.FileStorageRepository, config *entity.Config) usecase.FileUploadUseCase {
	return usecase.NewFileUploadUseCase(fileRepo, storageRepo, config.BaseStorageDir)
}

func FileDeleteUseCaseProvider(fileRepo db_repo.FileRepository, storageRepo fs_repo.FileStorageRepository, config *entity.Config) usecase.FileDeleteUseCase {
	return usecase.NewFileDeleteUseCase(config, fileRepo, storageRepo)
}

// ----------------------------------------------------------------
// Handler Providers
// ----------------------------------------------------------------
func FileUploadHandlerProvider(fileUploadUseCase usecase.FileUploadUseCase, config *entity.Config, logger *slog.Logger) *handler.FileUploadHandler {
	return handler.NewFileUploadHandler(fileUploadUseCase, config, logger)
}

func FileGetHandlerProvider(fileGetUseCase usecase.FileGetUseCase, config *entity.Config, logger *slog.Logger) *handler.FileGetHandler {
	return handler.NewFileGetHandler(fileGetUseCase, config, logger)
}

func FileDeleteHandlerProvider(fileDeleteUseCase usecase.FileDeleteUseCase, logger *slog.Logger) *handler.FileDeleteHandler {
	return handler.NewFileDeleteHandler(fileDeleteUseCase, logger)
}

// RouterProvider provides the router implementation
func RouterProvider(
	fileUploadHandler *handler.FileUploadHandler,
	fileGetHandler *handler.FileGetHandler,
	fileDeleteHandler *handler.FileDeleteHandler,
) *router.Router {
	r := router.NewRouter(fileUploadHandler, fileGetHandler, fileDeleteHandler)
	r.SetupRoutes()
	return r
}

// ServerProvider provides sets for dependency injection
var ServerProvider = wire.NewSet(
	ConfigProvider,
	LoggerProvider,
	DBProvider,
	// Repository
	FileRepositoryProvider,
	StorageRepositoryProvider,
	// Usecase
	FileGetUseCaseProvider,
	FileUploadUseCaseProvider,
	FileDeleteUseCaseProvider,
	// Handler
	FileUploadHandlerProvider,
	FileGetHandlerProvider,
	FileDeleteHandlerProvider,
	// Router
	RouterProvider,
)
