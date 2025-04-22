package di

import (
	"os"

	"github.com/google/wire"
	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/interface/command"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
	"golang.org/x/exp/slog"
)

// Command type aliases to distinguish between different commands
type UploadCommand *cobra.Command
type DeleteCommand *cobra.Command
type ListCommand *cobra.Command

// ConfigProvider provides the application configuration
func ConfigProvider() *entity.Config {
	return entity.NewServiceConfig()
}

// LoggerProvider provides the application logger
func LoggerProvider() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}

// ----------------------------------------------------------------
// Usecase Providers
// ----------------------------------------------------------------
func InitUploadUsecaseProvider(config *entity.Config) *usecase.InitUploadUsecase {
	return usecase.NewInitUploadUsecase(config)
}

func UploadUsecaseProvider(config *entity.Config) *usecase.UploadUsecase {
	return usecase.NewUploadUsecase(config)
}

func DeleteFileUsecaseProvider(config *entity.Config) *usecase.DeleteUsecase {
	return usecase.NewDeleteUsecase(config)
}

func ListUsecaseProvider(config *entity.Config) *usecase.ListUsecase {
	return usecase.NewListUsecase(config)
}

// UploadCommandProvider provides the upload command
func UploadCommandProvider(config *entity.Config, initUploadUsecase *usecase.InitUploadUsecase, uploadUsecase *usecase.UploadUsecase, deleteFileUsecase *usecase.DeleteUsecase) UploadCommand {
	return UploadCommand(command.NewUploadCommandHandler(config, initUploadUsecase, uploadUsecase, deleteFileUsecase).Execute())
}

// DeleteCommandProvider provides the delete command
func DeleteCommandProvider(config *entity.Config, deleteUsecase *usecase.DeleteUsecase) DeleteCommand {
	return DeleteCommand(command.NewDeleteCommandHandler(config, deleteUsecase).Execute())
}

// ListCommandProvider provides the list command
func ListCommandProvider(config *entity.Config, listUsecase *usecase.ListUsecase) ListCommand {
	return ListCommand(command.NewListCommandHandler(config, listUsecase).Execute())
}

// RootCommandProvider provides the root command with all subcommands
func RootCommandProvider(uploadCmd UploadCommand, deleteCmd DeleteCommand, listCmd ListCommand) *cobra.Command {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "fs-store",
		Short: "A command-line interface for the file storage server",
		Long:  `A command-line interface for interacting with the file storage server. It allows you to upload, delete, and list files.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().String("server", "http://localhost:38080", "Server origin URL")
	rootCmd.PersistentFlags().Int("concurrency", 5, "Maximum number of concurrent operations")

	// Add commands
	rootCmd.AddCommand((*cobra.Command)(uploadCmd))
	rootCmd.AddCommand((*cobra.Command)(deleteCmd))
	rootCmd.AddCommand((*cobra.Command)(listCmd))

	return rootCmd
}

// CLIProvider provides sets for dependency injection
var CLIProvider = wire.NewSet(
	ConfigProvider,
	LoggerProvider,
	// Usecase providers
	InitUploadUsecaseProvider,
	UploadUsecaseProvider,
	DeleteFileUsecaseProvider,
	ListUsecaseProvider,
	// Command providers
	UploadCommandProvider,
	DeleteCommandProvider,
	ListCommandProvider,
	RootCommandProvider,
)
