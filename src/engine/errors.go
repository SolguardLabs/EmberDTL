package engine

import "fmt"

type ErrorKind string

const (
	ErrInvalid  ErrorKind = "invalid"
	ErrNotFound ErrorKind = "not_found"
	ErrPolicy   ErrorKind = "policy"
	ErrState    ErrorKind = "state"
	ErrRisk     ErrorKind = "risk"
	ErrSolvency ErrorKind = "solvency"
)

type EngineError struct {
	Kind    ErrorKind
	Message string
}

func (e EngineError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func invalid(format string, args ...any) error {
	return EngineError{Kind: ErrInvalid, Message: fmt.Sprintf(format, args...)}
}

func notFound(format string, args ...any) error {
	return EngineError{Kind: ErrNotFound, Message: fmt.Sprintf(format, args...)}
}

func policyError(format string, args ...any) error {
	return EngineError{Kind: ErrPolicy, Message: fmt.Sprintf(format, args...)}
}

func stateError(format string, args ...any) error {
	return EngineError{Kind: ErrState, Message: fmt.Sprintf(format, args...)}
}

func riskError(format string, args ...any) error {
	return EngineError{Kind: ErrRisk, Message: fmt.Sprintf(format, args...)}
}

func solvencyError(format string, args ...any) error {
	return EngineError{Kind: ErrSolvency, Message: fmt.Sprintf(format, args...)}
}
