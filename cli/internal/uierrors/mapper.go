package uierrors

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"syscall"
)

// Map translates a standard Go error into a user-friendly UserError.
func Map(err error) *UserError {
	if err == nil {
		return nil
	}

	// If it's already a UserError, return it
	var userErr *UserError
	if errors.As(err, &userErr) {
		return userErr
	}

	// Network Errors
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if isConnectionRefused(err) {
			return New(
				"Connection Failed",
				"Could not connect to the GoConnect server.",
				"Ensure the server is running and the URL is correct in your config.",
				err,
			)
		}
		if netErr.Timeout() {
			return New(
				"Connection Timeout",
				"The server took too long to respond.",
				"Check your internet connection or firewall settings.",
				err,
			)
		}
	}

	// Context Errors
	if errors.Is(err, context.DeadlineExceeded) {
		return New(
			"Timeout",
			"The operation timed out.",
			"The server or peer might be busy or unreachable.",
			err,
		)
	}

	if errors.Is(err, context.Canceled) {
		return New(
			"Cancelled",
			"The operation was cancelled.",
			"",
			err,
		)
	}

	// File System Errors
	if errors.Is(err, os.ErrPermission) {
		return New(
			"Permission Denied",
			"You do not have permission to access a required file.",
			"Try running the command with 'sudo' or as an administrator.",
			err,
		)
	}

	if errors.Is(err, os.ErrNotExist) {
		return New(
			"Not Found",
			"A required file or directory was not found.",
			"Run 'goconnect doctor' to check your installation.",
			err,
		)
	}

	// Common Strings (fallback for un-typed errors)
	msg := err.Error()
	if strings.Contains(msg, "token expired") || strings.Contains(msg, "unauthorized") {
		return New(
			"Authentication Failed",
			"Your session has expired or is invalid.",
			"Run 'goconnect login' to authenticate again.",
			err,
		)
	}

	// Generic Fallback
	return New(
		"Error",
		"An unexpected error occurred.",
		fmt.Sprintf("Technical details: %s", err.Error()),
		err,
	)
}

func isConnectionRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(err.Error(), "connection refused")
}
