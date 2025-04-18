package di

import (
	"os"

	"github.com/google/wire"
	"github.com/spf13/cobra"
	"github.com/tomoya.tokunaga/cli/internal/interface/command"
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

// UploadCommandProvider provides the upload command
func UploadCommandProvider() UploadCommand {
	return UploadCommand(command.UploadCommand())
}

// DeleteCommandProvider provides the delete command
func DeleteCommandProvider() DeleteCommand {
	return DeleteCommand(command.DeleteCommand())
}

// ListCommandProvider provides the list command
func ListCommandProvider() ListCommand {
	return ListCommand(command.ListCommand())
}

// RootCommandProvider provides the root command with all subcommands
func RootCommandProvider(
	uploadCmd UploadCommand,
	deleteCmd DeleteCommand,
	listCmd ListCommand,
) *cobra.Command {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "fs-store",
		Short: "A command-line interface for the file storage server",
		Long:  `A command-line interface for interacting with the file storage server. It allows you to upload, delete, and list files.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().String("server", "http://localhost:18080", "Server origin URL")
	rootCmd.PersistentFlags().Int("concurrency", 5, "Maximum number of concurrent operations")

	// Add commands
	rootCmd.AddCommand((*cobra.Command)(uploadCmd))
	rootCmd.AddCommand((*cobra.Command)(deleteCmd))
	rootCmd.AddCommand((*cobra.Command)(listCmd))

	return rootCmd
}

// CLIProvider provides sets for dependency injection
var CLIProvider = wire.NewSet(
	LoggerProvider,
	UploadCommandProvider,
	DeleteCommandProvider,
	ListCommandProvider,
	RootCommandProvider,
)
