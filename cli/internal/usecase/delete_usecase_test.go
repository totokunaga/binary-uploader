package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomoya.tokunaga/cli/internal/mock"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestDeleteUsecase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		targetFileName string
		mockSetup      func(mockClient *mock.MockFileServerHttpClient)
		expectedErr    error
	}{
		{
			name:           "successful deletion",
			targetFileName: "test-file.txt",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					DeleteFile(gomock.Any(), "test-file.txt").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:           "error from file server",
			targetFileName: "non-existent-file.txt",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					DeleteFile(gomock.Any(), "non-existent-file.txt").
					Return(errors.New("file not found"))
			},
			expectedErr: errors.New("failed to delete file 'non-existent-file.txt': file not found"),
		},
		{
			name:           "empty file name",
			targetFileName: "",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					DeleteFile(gomock.Any(), "").
					Return(errors.New("file name cannot be empty"))
			},
			expectedErr: errors.New("failed to delete file '': file name cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockFileServerHttpClient(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			usecase := usecase.NewDeleteUsecase(mockClient)
			err := usecase.Execute(context.Background(), tt.targetFileName)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
