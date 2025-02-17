package avm

import (
	"fmt"
	"strings"
)

// TxnConfirmationErrorType represents the type of error waiting for a txn confirmation
type TxnConfirmationErrorType int

const (
	ErrWaitTimeout TxnConfirmationErrorType = iota
	ErrRejected
	ErrInternal
)

func (e TxnConfirmationErrorType) String() string {
	switch e {
	case ErrWaitTimeout:
		return "TxnConfirmationTimeoutError"
	case ErrRejected:
		return "TxnConfirmationRejectionError"
	case ErrInternal:
		return "TxnConfirmationInternalError"
	default:
		return "TxnConfirmationUnknownError"
	}
}

// TxnConfirmationError represents an error waiting for a txn confirmation
type TxnConfirmationError struct {
	Type    TxnConfirmationErrorType // The type of the error
	Message string                   // The original error message
}

// Implement the Error() method to satisfy the error interface
func (e *TxnConfirmationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// parseWaitForConfirmationError parses the error returned by WaitForConfirmation
func parseWaitForConfirmationError(err error) *TxnConfirmationError {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "timed out") {
		return &TxnConfirmationError{
			Type:    ErrWaitTimeout,
			Message: err.Error(),
		}
	}
	if strings.Contains(err.Error(), "Transaction rejected") {
		return &TxnConfirmationError{
			Type:    ErrRejected,
			Message: err.Error(),
		}
	}
	return &TxnConfirmationError{
		Type:    ErrInternal,
		Message: err.Error(),
	}
}

func InternalError(s string) *TxnConfirmationError {
	return &TxnConfirmationError{
		Type:    ErrInternal,
		Message: s,
	}
}
