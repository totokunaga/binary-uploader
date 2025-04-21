package di

import (
	"os"

	"golang.org/x/exp/slog"
	"gorm.io/gorm"

	"github.com/google/wire"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	"github.com/tomoya.tokunaga/server/internal/domain/repository"
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

// DBProvider provides the database connection
func DBProvider(config *entity.Config, logger *slog.Logger) (*gorm.DB, error) {
	return database.NewDB(config, logger)
}

// ----------------------------------------------------------------
// Repository Providers
// ----------------------------------------------------------------

func FileRepositoryProvider(db *gorm.DB) repository.FileRepository {
	return db_repo.NewFileRepository(db)
}

func FileChunkRepositoryProvider(db *gorm.DB) repository.FileChunkRepository {
	return db_repo.NewFileChunkRepository(db)
}

func StorageRepositoryProvider() repository.StorageRepository {
	return fs_repo.NewStorageRepository()
}

// ----------------------------------------------------------------
// Usecase Providers
// ----------------------------------------------------------------

func FileListUseCaseProvider(fileRepo repository.FileRepository) usecase.FileListUseCase {
	return usecase.NewFileListUseCase(fileRepo)
}

func FileUploadUseCaseProvider(fileRepo repository.FileRepository, fileChunkRepo repository.FileChunkRepository, storageRepo repository.StorageRepository, config *entity.Config) usecase.FileUploadUseCase {
	return usecase.NewFileUploadUseCase(fileRepo, fileChunkRepo, storageRepo, config.BaseStorageDir, config.UploadSizeLimit)
}

func FileDeleteUseCaseProvider(fileRepo repository.FileRepository, fileChunkRepo repository.FileChunkRepository, storageRepo repository.StorageRepository, config *entity.Config) usecase.FileDeleteUseCase {
	return usecase.NewFileDeleteUseCase(config, fileRepo, fileChunkRepo, storageRepo)
}

// ----------------------------------------------------------------
// Handler Providers
// ----------------------------------------------------------------

func FileUploadHandlerProvider(fileUploadUseCase usecase.FileUploadUseCase, logger *slog.Logger) *handler.FileUploadHandler {
	return handler.NewFileUploadHandler(fileUploadUseCase, logger)
}

func FileListHandlerProvider(fileListUseCase usecase.FileListUseCase, logger *slog.Logger) *handler.FileListHandler {
	return handler.NewFileListHandler(fileListUseCase, logger)
}

func FileDeleteHandlerProvider(fileDeleteUseCase usecase.FileDeleteUseCase, logger *slog.Logger) *handler.FileDeleteHandler {
	return handler.NewFileDeleteHandler(fileDeleteUseCase, logger)
}

// RouterProvider provides the router implementation
func RouterProvider(
	fileUploadHandler *handler.FileUploadHandler,
	fileListHandler *handler.FileListHandler,
	fileDeleteHandler *handler.FileDeleteHandler,
) *router.Router {
	r := router.NewRouter(fileUploadHandler, fileListHandler, fileDeleteHandler)
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
	FileChunkRepositoryProvider,
	StorageRepositoryProvider,
	// Usecase
	FileListUseCaseProvider,
	FileUploadUseCaseProvider,
	FileDeleteUseCaseProvider,
	// Handler
	FileUploadHandlerProvider,
	FileListHandlerProvider,
	FileDeleteHandlerProvider,
	// Router
	RouterProvider,
)
