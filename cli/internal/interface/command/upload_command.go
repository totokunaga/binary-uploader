package command

import (
	"fmt"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
)

type UploadCommandHandler struct {
	config            *entity.Config
	initUploadUsecase *usecase.InitUploadUsecase
	uploadUsecase     *usecase.UploadUsecase
	deleteUsecase     *usecase.DeleteUsecase
}

func NewUploadCommandHandler(
	config *entity.Config,
	initUploadUsecase *usecase.InitUploadUsecase,
	uploadUsecase *usecase.UploadUsecase,
	deleteUsecase *usecase.DeleteUsecase,
) *UploadCommandHandler {
	return &UploadCommandHandler{
		config:            config,
		initUploadUsecase: initUploadUsecase,
		uploadUsecase:     uploadUsecase,
		deleteUsecase:     deleteUsecase,
	}
}

func (h *UploadCommandHandler) Execute() *cobra.Command {
	var (
		retries   int
		chunkSize int64
		fileName  string
	)

	cmd := &cobra.Command{
		Use:   "upload-file [file path]",
		Short: "Upload a file to the file server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// There's no need to check the existence of the first argument because cobra.ExactArgs(1) handles it
			filePath := args[0]

			// Uses the user-preferred file name for this upload if specified
			targetFileName := fileName
			if targetFileName == "" {
				targetFileName = filepath.Base(filePath)
			}

			// Checks if there's a file with the same name on the server
			postPrecheckAction, precheckOutput, err := h.initUploadUsecase.ExecutePrecheck(filePath, targetFileName)
			if err != nil {
				return fmt.Errorf("[ERROR] failed to initialize upload: %w", err)
			}

			// Determines if this is a re-upload based on the result of precheck usecase
			isReUpload := postPrecheckAction == usecase.ProceedWithReUpload

			// Initializes the upload usecase
			uploadInitOutput := &usecase.UploadUsecaseOutput{}

			// Takes an appropriate action based on the result of precheck usecase
			switch postPrecheckAction {
			case usecase.Exits:
				// When the file with a same name and same contents exist in the file server
				fmt.Printf("'%s' already uploaded to the file server.\n", targetFileName)
				fmt.Printf("Exiting...\n")
				return nil
			case usecase.ProceedWithInit, usecase.ProceedWithReUpload:
				// When there's no conflicting file exists on the file server, or the file with a same name and same contents were tried to be
				// uploaded but encoutered some problems before. Sends the upload-init request
				uploadInitOutput, err = h.initUploadUsecase.Execute(filePath, targetFileName, precheckOutput.Checksum, chunkSize, isReUpload)
				if err != nil {
					return fmt.Errorf("[ERROR] failed to initialize upload: %w", err)
				}
			case usecase.SuggestExistingEntryDeletion:
				// When there's a conflicting file exists on the file server (e.g. a file with the same name but different contents)
				// Asks the user if they want to delete the conflicting file on the file server and retry the upload
				forceDeleted, err := h.handleFileConflict(targetFileName)
				if err != nil {
					return err
				}
				if !forceDeleted {
					fmt.Println("Cancelling the upload...")
					return nil
				}
				// Sends the upload-init request for the target file (the rest of the process is same as ProceedWithInit and ProceedWithReUpload)
				uploadInitOutput, err = h.initUploadUsecase.Execute(filePath, targetFileName, precheckOutput.Checksum, chunkSize, isReUpload)
				if err != nil {
					return fmt.Errorf("[ERROR] failed to initialize upload: %w", err)
				}
			default:
				return fmt.Errorf("[ERROR] unexpected post-precheck action: %v", postPrecheckAction)
			}

			// Calculates the number of bytes to be uploaded to the file server. If it's re-uploading, it calculates based on the number of
			// missing bytes reported by the file server in the upload-init response.
			var uploadChunkSizeTotal int64
			if isReUpload {
				uploadChunkSizeTotal = int64(uploadInitOutput.UploadChunkSize * uint64(len(uploadInitOutput.MissingChunkNumberMap)))
			} else {
				uploadChunkSizeTotal = precheckOutput.FileSize
			}

			// Sets up the progress bar
			bar := progressbar.NewOptions64(
				uploadChunkSizeTotal,
				progressbar.OptionSetDescription("Uploading in progress..."),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(30),
				progressbar.OptionThrottle(100),
				progressbar.OptionShowCount(),
				progressbar.OptionFullWidth(),
				progressbar.OptionSetRenderBlankState(true),
			)

			// Uploads the file to the file server chunk by chunk
			err = h.uploadUsecase.Execute(&usecase.UploadUsecaseInput{
				UploadID:              uploadInitOutput.UploadID,
				FilePath:              filePath,
				IsReUpload:            isReUpload,
				MissingChunkNumberMap: uploadInitOutput.MissingChunkNumberMap,
				ProgressCb:            func(size int64) { _ = bar.Add64(size) },
			})
			if err != nil {
				deleteErr := h.deleteUsecase.Execute(targetFileName)
				if deleteErr != nil {
					return fmt.Errorf("[ERROR] %w", deleteErr)
				}
				return fmt.Errorf("upload failed for file '%s': %w", targetFileName, err)
			}

			fmt.Printf("\nSuccessfully uploaded!")
			return nil
		},
	}

	cmd.Flags().IntVarP(&retries, "retries", "r", entity.DefaultRetries, "Number of retries for failed chunk uploads")
	cmd.Flags().Int64VarP(&chunkSize, "chunk-size", "c", entity.DefaultChunkSize, "Chunk size in bytes")
	cmd.Flags().StringVarP(&fileName, "file-name", "n", "", "Specify the file name to be used on the server")

	return cmd
}

func (h *UploadCommandHandler) handleFileConflict(targetFileName string) (bool, error) {
	// Asks the user if they want to delete the conflicting file on the file server and retry the upload
	fmt.Printf("[ERROR] A confilcting file (a file with the same name but different contents) is found on the file server\n")
	fmt.Printf("* Do you want to delete the conflicting file and proceed with the upload? (y/n): ")

	// Reads the user input
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	fmt.Println()

	if input == "y" || input == "Y" {
		fmt.Printf("Startting to delete the conflicting file \"%s\" on the file server...\n", targetFileName)
		if delErr := h.deleteUsecase.Execute(targetFileName); delErr != nil {
			return false, fmt.Errorf("[ERROR] failed to delete the conflicting file \"%s\": %w", targetFileName, delErr)
		}
		fmt.Printf("Successfully deleted the conflicting file\n\n")
		return true, nil
	}
	return false, nil
}
