package uierrors

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name          string
		input         error
		expectedTitle string
	}{
		{
			name:          "Nil Error",
			input:         nil,
			expectedTitle: "",
		},
		{
			name:          "Connection Refused",
			input:         &net.OpError{Err: syscall.ECONNREFUSED},
			expectedTitle: "Connection Failed",
		},
		{
			name:          "Context Timeout",
			input:         context.DeadlineExceeded,
			expectedTitle: "Timeout",
		},
		{
			name:          "Permission Denied",
			input:         os.ErrPermission,
			expectedTitle: "Permission Denied",
		},
		{
			name:          "File Not Found",
			input:         os.ErrNotExist,
			expectedTitle: "Not Found",
		},
		{
			name:          "Mapped UserError (Idempotency)",
			input:         New("Custom", "Msg", "Hint", nil),
			expectedTitle: "Custom",
		},
		{
			name:          "Generic Error",
			input:         errors.New("something went wrong"),
			expectedTitle: "Error",
		},
		{
			name:          "Auth Error String",
			input:         errors.New("rpc error: code = Unauthenticated desc = token expired"),
			expectedTitle: "Authentication Failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.input)
			if tt.input == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTitle, result.Title)
			}
		})
	}
}

func TestUserError_Error(t *testing.T) {
	err := New("Title", "Message", "Hint", nil)
	expected := "Title: Message (Hint: Hint)"
	assert.Equal(t, expected, err.Error())

	errNoHint := New("Title", "Message", "", nil)
	expectedNoHint := "Title: Message"
	assert.Equal(t, expectedNoHint, errNoHint.Error())
}
