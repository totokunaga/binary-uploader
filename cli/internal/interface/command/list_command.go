package command

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

// ListCommand creates a command to list files on the server
func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-files",
		Short: "List files available on the file server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create list service
			listService := usecase.NewListService(cmd)

			// Execute file listing
			files, err := listService.ListFiles()
			if err != nil {
				return fmt.Errorf("failed to list files: %w", err)
			}

			// Format and display the results
			if len(files) == 0 {
				fmt.Println("No files found on the server.")
				return nil
			}

			// Create a tabwriter for aligned output
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tSIZE\tCREATED AT")
			fmt.Fprintln(w, "----\t----\t----------")

			for _, file := range files {
				// Format file size for human readability
				sizeStr := formatFileSize(file.Size)

				// Format timestamp
				timeStr := file.CreatedAt.Format(time.RFC3339)

				fmt.Fprintf(w, "%s\t%s\t%s\n", file.Name, sizeStr, timeStr)
			}

			w.Flush()
			return nil
		},
	}

	return cmd
}

// formatFileSize formats file size in bytes to human-readable format
func formatFileSize(sizeInBytes int64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota)
		MB
		GB
		TB
	)

	var unit string
	var value float64

	switch {
	case sizeInBytes >= TB:
		unit = "TB"
		value = float64(sizeInBytes) / TB
	case sizeInBytes >= GB:
		unit = "GB"
		value = float64(sizeInBytes) / GB
	case sizeInBytes >= MB:
		unit = "MB"
		value = float64(sizeInBytes) / MB
	case sizeInBytes >= KB:
		unit = "KB"
		value = float64(sizeInBytes) / KB
	default:
		unit = "B"
		value = float64(sizeInBytes)
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}
