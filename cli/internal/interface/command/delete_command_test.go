package command_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/tomoya.tokunaga/cli/internal/interface/command"
	"github.com/tomoya.tokunaga/cli/internal/mock"
)

func TestDeleteCommandHandler_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		args          []string
		mockSetup     func(m *mock.MockDeleteUsecase)
		expectedOut   string
		expectedError error
	}{
		{
			name: "Success: Delete file successfully",
			args: []string{"testfile.txt"},
			mockSetup: func(m *mock.MockDeleteUsecase) {
				m.EXPECT().Execute(ctx, "testfile.txt").Return(nil)
			},
			expectedOut: `Deleting 'testfile.txt'...
Successfully deleted!
`,
		},
		{
			name: "Success: Delete file with path successfully",
			args: []string{"/path/to/testfile.txt"},
			mockSetup: func(m *mock.MockDeleteUsecase) {
				m.EXPECT().Execute(ctx, "testfile.txt").Return(nil)
			},
			expectedOut: `Deleting 'testfile.txt'...
Successfully deleted!
`,
		},
		{
			name:          "Error: Invalid file name (empty after base)",
			args:          []string{"/"},
			mockSetup:     func(m *mock.MockDeleteUsecase) {},
			expectedError: errors.New("[ERROR] Invalid file name: /. Please provide a valid file name"),
			expectedOut: `Error: [ERROR] Invalid file name: /. Please provide a valid file name
Usage:
  delete-file [file name] [flags]

Flags:
  -h, --help   help for delete-file

`,
		},
		{
			name: "Error: Usecase returns error",
			args: []string{"errorfile.txt"},
			mockSetup: func(m *mock.MockDeleteUsecase) {
				m.EXPECT().Execute(ctx, "errorfile.txt").Return(errors.New("delete failed"))
			},
			expectedOut: `Deleting 'errorfile.txt'...
[ERROR] Delete 'errorfile.txt' failed: delete failed
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDeleteUsecase := mock.NewMockDeleteUsecase(ctrl)
			tc.mockSetup(mockDeleteUsecase)

			handler := command.NewDeleteCommandHandler(mockDeleteUsecase)
			cmd := handler.Execute()

			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedOut, out.String())
		})
	}
}
