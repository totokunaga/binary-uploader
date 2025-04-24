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

func TestListUsecase_Execute(t *testing.T) {
	tests := []struct {
		name              string
		mockSetup         func(mockClient *mock.MockFileServerHttpClient)
		expectedOutput    *entity.ListFilesResp
		expectedErr       bool
		expectedErrString string
	}{
		{
			name: "When list API returns success, should return file list",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					ListFiles(gomock.Any()).
					Return(&entity.ListFilesResp{
						Files: []string{
							"test1.txt",
							"test2.txt",
						},
					}, nil)
			},
			expectedOutput: &entity.ListFilesResp{
				Files: []string{
					"test1.txt",
					"test2.txt",
				},
			},
			expectedErr: false,
		},
		{
			name: "When list API returns empty list, should return empty list",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					ListFiles(gomock.Any()).
					Return(&entity.ListFilesResp{
						Files: []string{},
					}, nil)
			},
			expectedOutput: &entity.ListFilesResp{
				Files: []string{},
			},
			expectedErr: false,
		},
		{
			name: "When list API returns error, should return error",
			mockSetup: func(mockClient *mock.MockFileServerHttpClient) {
				mockClient.EXPECT().
					ListFiles(gomock.Any()).
					Return(nil, errors.New("API error"))
			},
			expectedOutput:    nil,
			expectedErr:       true,
			expectedErrString: "failed to list files: API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockFileServerHttpClient(ctrl)
			tt.mockSetup(mockClient)

			usecase := usecase.NewListUsecase(mockClient)
			output, err := usecase.Execute(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedOutput.Files), len(output.Files))

				for i, expectedFile := range tt.expectedOutput.Files {
					assert.Equal(t, expectedFile, output.Files[i])
				}
			}
		})
	}
}
