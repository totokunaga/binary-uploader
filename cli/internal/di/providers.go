package di

import (
	"os"

	"github.com/google/wire"
	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/interface/command"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
	"golang.org/x/exp/slog"
)

// Command type aliases to distinguish between different commands
type UploadCommand *cobra.Command
type DeleteCommand *cobra.Command
type ListCommand *cobra.Command

// LoggerProvider provides the application logger
func LoggerProvider() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}

// ----------------------------------------------------------------
// Usecase Providers
// ----------------------------------------------------------------
func InitUploadUsecaseProvider() *usecase.InitUploadUsecase {
	return usecase.NewInitUploadUsecase()
}

func UploadUsecaseProvider() *usecase.UploadUsecase {
	return usecase.NewUploadUsecase()
}

func DeleteFileUsecaseProvider() *usecase.DeleteUsecase {
	return usecase.NewDeleteUsecase()
}

func ListUsecaseProvider() *usecase.ListUsecase {
	return usecase.NewListUsecase()
}

// UploadCommandProvider provides the upload command
func UploadCommandProvider(initUploadUsecase *usecase.InitUploadUsecase, uploadUsecase *usecase.UploadUsecase, deleteFileUsecase *usecase.DeleteUsecase) UploadCommand {
	return UploadCommand(command.NewUploadCommandHandler(initUploadUsecase, uploadUsecase, deleteFileUsecase).Execute())
}

// DeleteCommandProvider provides the delete command
func DeleteCommandProvider(deleteUsecase *usecase.DeleteUsecase) DeleteCommand {
	return DeleteCommand(command.NewDeleteCommandHandler(deleteUsecase).Execute())
}

// ListCommandProvider provides the list command
func ListCommandProvider(listUsecase *usecase.ListUsecase) ListCommand {
	return ListCommand(command.NewListCommandHandler(listUsecase).Execute())
}

// RootCommandProvider provides the root command with all subcommands
func RootCommandProvider(uploadCmd UploadCommand, deleteCmd DeleteCommand, listCmd ListCommand) *cobra.Command {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "fs-store",
		Short: "Command-line interface for the file storage server",
		Long:  `Command-line interface for interacting with the file storage server to upload, delete, and list files.`,
	}

	// Add commands
	rootCmd.AddCommand((*cobra.Command)(uploadCmd))
	rootCmd.AddCommand((*cobra.Command)(deleteCmd))
	rootCmd.AddCommand((*cobra.Command)(listCmd))

	return rootCmd
}

// CLIProvider provides sets for dependency injection
var CLIProvider = wire.NewSet(
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
