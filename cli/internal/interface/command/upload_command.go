package command

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

type UploadCommandHandler struct {
	initUploadUsecase usecase.InitUploadUsecase
	uploadUsecase     usecase.UploadUsecase
	deleteUsecase     usecase.DeleteUsecase
}

func NewUploadCommandHandler(
	initUploadUsecase usecase.InitUploadUsecase,
	uploadUsecase usecase.UploadUsecase,
	deleteUsecase usecase.DeleteUsecase,
) *UploadCommandHandler {
	return &UploadCommandHandler{
		initUploadUsecase: initUploadUsecase,
		uploadUsecase:     uploadUsecase,
		deleteUsecase:     deleteUsecase,
	}
}

func (h *UploadCommandHandler) Execute() *cobra.Command {
	var (
		concurrency        int
		retries            int
		chunkSize          int64
		fileName           string
		compressionEnabled bool
	)

	cmd := &cobra.Command{
		Use:   "upload-file [file path]",
		Short: "Upload a file to the file server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			ctx := cmd.Context()

			// Add command parameters to context
			ctx = context.WithValue(ctx, entity.ConcurrencyKey, concurrency)
			ctx = context.WithValue(ctx, entity.RetriesKey, retries)
			ctx = context.WithValue(ctx, entity.CompressionEnabledKey, compressionEnabled)

			targetFileName := fileName
			if targetFileName == "" {
				targetFileName = filepath.Base(filePath)
			}

			// Precheck
			postPrecheckAction, precheckOutput, err := h.initUploadUsecase.ExecutePrecheck(ctx, &usecase.InitUploadPrecheckUsecaseInput{
				FilePath:       filePath,
				TargetFileName: targetFileName,
			})
			if err != nil {
				cmd.PrintErrf("[ERROR] Failed to initialize upload pre-check: %v\n", err)
				return nil
			}

			isReUpload := postPrecheckAction == usecase.ProceedWithReUpload
			uploadInitOutput := &usecase.UploadUsecaseOutput{}

			// Handle precheck results
			switch postPrecheckAction {
			case usecase.Exits:
				cmd.Printf("'%s' already uploaded to the file server.\n", targetFileName)
				cmd.Println("Exiting...")
				return nil
			case usecase.ProceedWithInit, usecase.ProceedWithReUpload:
				uploadInitOutput, err = h.initUploadUsecase.Execute(ctx, &usecase.InitUploadUsecaseInput{
					FilePath:         filePath,
					TargetFileName:   targetFileName,
					OriginalChecksum: precheckOutput.Checksum,
					ChunkSize:        chunkSize,
					IsReUpload:       isReUpload,
				})
				if err != nil {
					// Return error here because it's a fundamental failure of init
					return fmt.Errorf("[ERROR] failed to initialize upload: %w", err)
				}
			case usecase.SuggestExistingEntryDeletion:
				forceDeleted, err := h.handleFileConflict(cmd, ctx, targetFileName)
				if err != nil {
					cmd.PrintErrf("[ERROR] Failed to handle file conflict: %v\n", err)
					return nil
				}
				if !forceDeleted {
					cmd.Println("Cancelling the upload...")
					return nil
				}
				// Re-init after deletion
				uploadInitOutput, err = h.initUploadUsecase.Execute(ctx, &usecase.InitUploadUsecaseInput{
					FilePath:         filePath,
					TargetFileName:   targetFileName,
					OriginalChecksum: precheckOutput.Checksum,
					ChunkSize:        chunkSize,
					IsReUpload:       false, // It's a fresh upload after deletion
				})
				if err != nil {
					cmd.PrintErrf("[ERROR] Failed to initialize upload after conflict resolution: %v\n", err)
					return nil
				}
			default:
				cmd.PrintErrf("[ERROR] Unexpected post-precheck action: %v\n", postPrecheckAction)
				return nil
			}

			// Calculate total size for progress bar
			var uploadChunkSizeTotal int64
			if isReUpload && uploadInitOutput.MissingChunkNumberMap != nil {
				uploadChunkSizeTotal = int64(uploadInitOutput.UploadChunkSize * uint64(len(uploadInitOutput.MissingChunkNumberMap)))
			} else {
				uploadChunkSizeTotal = precheckOutput.FileSize
			}

			// Setup progress bar to write to cmd's error stream
			bar := progressbar.NewOptions64(
				uploadChunkSizeTotal,
				progressbar.OptionSetDescription("Uploading in progress..."),
				progressbar.OptionSetWriter(cmd.ErrOrStderr()),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(30),
				progressbar.OptionThrottle(100),
				progressbar.OptionShowCount(),
				progressbar.OptionFullWidth(),
				progressbar.OptionSetRenderBlankState(true),
				progressbar.OptionClearOnFinish(),
			)

			// Execute upload
			err = h.uploadUsecase.Execute(ctx, &usecase.UploadUsecaseInput{
				UploadID:              uploadInitOutput.UploadID,
				FilePath:              filePath,
				ChunkSize:             chunkSize,
				IsReUpload:            isReUpload,
				MissingChunkNumberMap: uploadInitOutput.MissingChunkNumberMap,
				ProgressCb:            func(size int64) { _ = bar.Add64(size) },
			})
			if err != nil {
				_ = bar.Clear()
				deleteErr := h.deleteUsecase.Execute(ctx, targetFileName)
				if deleteErr != nil {
					cmd.PrintErrf("[ERROR] Failed to delete the partially uploaded file entry: %v\n", deleteErr)
					cmd.PrintErrf("[ERROR] Upload failed for file '%s': %v\n", targetFileName, err)
					return nil
				}
				cmd.PrintErrf("[ERROR] Upload failed for file '%s': %v\n", targetFileName, err)
				return nil
			}

			cmd.Println("Successfully uploaded!")
			return nil
		},
	}

	// Define flags
	cmd.Flags().IntVarP(&retries, "retries", "r", entity.DefaultRetries, "Number of retries for failed chunk uploads")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", entity.DefaultMaxConcurrency, "Maximum number of concurrent operations")
	cmd.Flags().Int64VarP(&chunkSize, "chunk-size", "s", entity.DefaultChunkSize, "Chunk size in bytes")
	cmd.Flags().StringVarP(&fileName, "file-name", "n", "", "Specify the file name to be used on the server")
	cmd.Flags().BoolVarP(&compressionEnabled, "compression", "z", entity.DefaultCompressionEnabled, "Enable gzip compression for the file (data will be decompressed on the file server)")

	return cmd
}

// handleFileConflict now accepts *cobra.Command to use its I/O streams
func (h *UploadCommandHandler) handleFileConflict(cmd *cobra.Command, ctx context.Context, targetFileName string) (bool, error) {
	cmd.PrintErrln("[ERROR] A confilcting file (a file with the same name but different contents) is found on the file server")
	cmd.PrintErrf("* Do you want to delete the conflicting file and proceed with the upload? (y/n): ")

	// Use bufio.Reader to read from cmd's input stream
	reader := bufio.NewReader(cmd.InOrStdin())
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	input = strings.TrimSpace(input)

	if strings.EqualFold(input, "y") {
		cmd.Printf("Attempting to delete the conflicting file \"%s\" on the file server...\n", targetFileName)
		if delErr := h.deleteUsecase.Execute(ctx, targetFileName); delErr != nil {
			cmd.PrintErrf("[ERROR] failed to delete the conflicting file \"%s\": %v\n", targetFileName, delErr)
			return false, fmt.Errorf("failed to delete conflicting file")
		}
		cmd.Println("Successfully deleted the conflicting file.")
		return true, nil
	}
	return false, nil
}
