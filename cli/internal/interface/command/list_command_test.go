package command_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/interface/command"
	"github.com/tomoya.tokunaga/cli/internal/mock"
)

func TestListCommandHandler_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		args          []string
		mockSetup     func(m *mock.MockListUsecase)
		expectedOut   string
		expectedError bool
	}{
		{
			name: "Success: List files successfully",
			args: []string{},
			mockSetup: func(m *mock.MockListUsecase) {
				m.EXPECT().Execute(ctx).Return(&entity.ListFilesResp{
					Files: []string{"file1.txt", "file2.jpg"},
				}, nil)
			},
			expectedOut: fmt.Sprintf("%v\n", []string{"file1.txt", "file2.jpg"}),
		},
		{
			name: "Error: Usecase returns error",
			args: []string{},
			mockSetup: func(m *mock.MockListUsecase) {
				m.EXPECT().Execute(ctx).Return(nil, errors.New("failed to fetch"))
			},
			expectedOut: "[ERROR] Failed to fetch files: failed to fetch\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockListUsecase := mock.NewMockListUsecase(ctrl)
			tc.mockSetup(mockListUsecase)

			handler := command.NewListCommandHandler(mockListUsecase)
			cmd := handler.Execute()

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOut, out.String())
		})
	}
}
