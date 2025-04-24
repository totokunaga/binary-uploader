package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/mock"
	"github.com/tomoya.tokunaga/cli/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestUploadUsecase_Execute(t *testing.T) {
	tests := []struct {
		name              string
		input             *usecase.UploadUsecaseInput
		mockSetup         func(mockClient *mock.MockFileServerHttpClient)
		expectedErr       bool
		expectedErrString string
	}{
		{
			name: "When upload chunks succeed, should return no error",
			input: &usecase.UploadUsecaseInput{
				UploadID:              123,
				FilePath:              "testdata/test.txt",
				ChunkSize:             1024,
				IsReUpload:            false,
				MissingChunkNumberMap: make(map[uint64]struct{}),
			},
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					UploadChunk(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "When upload chunk fails with retry exhausted, should return error",
			input: &usecase.UploadUsecaseInput{
				UploadID:              123,
				FilePath:              "testdata/test.txt",
				ChunkSize:             1024,
				IsReUpload:            false,
				MissingChunkNumberMap: make(map[uint64]struct{}),
			},
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					UploadChunk(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("chunk upload error")).
					AnyTimes()
			},
			expectedErr:       true,
			expectedErrString: "failed to upload chunk",
		},
		{
			name: "When is re-upload, should only upload missing chunks",
			input: &usecase.UploadUsecaseInput{
				UploadID:   123,
				FilePath:   "testdata/test.txt",
				ChunkSize:  1024,
				IsReUpload: true,
				MissingChunkNumberMap: map[uint64]struct{}{
					1: {},
					3: {},
				},
			},
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					UploadChunk(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.WithValue(context.Background(), entity.ConcurrencyKey, 2)
			ctx = context.WithValue(ctx, entity.RetriesKey, 1)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockFileServerHttpClient(ctrl)
			tt.mockSetup(mockClient)

			usecase := usecase.NewUploadUsecase(mockClient)
			err := usecase.Execute(ctx, tt.input)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
