package di

import (
	"os"

	"golang.org/x/exp/slog"
	"gorm.io/gorm"

	"github.com/google/wire"
	"github.com/tomoya.tokunaga/server/internal/core/entity"
	"github.com/tomoya.tokunaga/server/internal/core/repository"
	"github.com/tomoya.tokunaga/server/internal/infra/database"
	"github.com/tomoya.tokunaga/server/internal/interface/api/handler"
	"github.com/tomoya.tokunaga/server/internal/interface/api/router"
	mysqldriver "github.com/tomoya.tokunaga/server/internal/interface/database/mysql"
	"github.com/tomoya.tokunaga/server/internal/interface/storage/filesystem"
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

// DBProvider provides the database connection
func DBProvider(config *entity.Config, logger *slog.Logger) (*gorm.DB, error) {
	return database.NewDB(config, logger)
}

// FileRepositoryProvider provides the file repository implementation
func FileRepositoryProvider(db *gorm.DB) repository.FileRepository {
	return mysqldriver.NewFileRepository(db)
}

// FileChunkRepositoryProvider provides the file chunk repository implementation
func FileChunkRepositoryProvider(db *gorm.DB) repository.FileChunkRepository {
	return mysqldriver.NewFileChunkRepository(db)
}

// StorageRepositoryProvider provides the storage repository implementation
func StorageRepositoryProvider() repository.StorageRepository {
	return filesystem.NewStorageRepository()
}

// FileUseCaseProvider provides the file use case implementation
func FileUseCaseProvider(
	fileRepo repository.FileRepository,
	fileChunkRepo repository.FileChunkRepository,
	storageRepo repository.StorageRepository,
	config *entity.Config,
) usecase.FileUseCase {
	return usecase.NewFileUseCase(
		fileRepo,
		fileChunkRepo,
		storageRepo,
		config.BaseStorageDir,
		config.UploadSizeLimit,
	)
}

// FileHandlerProvider provides the file handler implementation
func FileHandlerProvider(fileUseCase usecase.FileUseCase, logger *slog.Logger) *handler.FileHandler {
	return handler.NewFileHandler(fileUseCase, logger)
}

// RouterProvider provides the router implementation
func RouterProvider(fileHandler *handler.FileHandler) *router.Router {
	r := router.NewRouter(fileHandler)
	r.SetupRoutes()
	return r
}

// ServerProvider provides sets for dependency injection
var ServerProvider = wire.NewSet(
	ConfigProvider,
	LoggerProvider,
	DBProvider,
	FileRepositoryProvider,
	FileChunkRepositoryProvider,
	StorageRepositoryProvider,
	FileUseCaseProvider,
	FileHandlerProvider,
	RouterProvider,
)
