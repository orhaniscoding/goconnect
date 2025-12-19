package uierrors

import "fmt"

// UserError represents an error that is safe to display to the user.
type UserError struct {
	Title   string // Short summary (e.g., "Connection Failed")
	Message string // Friendly description
	Hint    string // Actionable advice
	Original error // The original error for logging (optional)
}

func (e *UserError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (Hint: %s)", e.Title, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Title, e.Message)
}

// New creates a new UserError
func New(title, message, hint string, original error) *UserError {
	return &UserError{
		Title:    title,
		Message:  message,
		Hint:     hint,
		Original: original,
	}
}
