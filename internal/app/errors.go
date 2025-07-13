package app

import "fmt"

// ErrorType はエラーの種類を表します。
type ErrorType string

const (
	// ErrTypeValidation はバリデーションエラーを表します。
	ErrTypeValidation ErrorType = "Validation"
	// ErrTypeNotFound はリソースが見つからないエラーを表します。
	ErrTypeNotFound ErrorType = "NotFound"
	// ErrTypeIO はファイルI/Oエラーを表します。
	ErrTypeIO ErrorType = "IO"
	// ErrTypeInternal は予期せぬ内部エラーを表します。
	ErrTypeInternal ErrorType = "Internal"
)

// AppError はアプリケーション固有のエラーを表す構造体です。
type AppError struct {
	// Type はエラーの種類です。
	Type ErrorType
	// Message はユーザー向けのメッセージです。
	Message string
	// original は元のエラーです。
	original error
}

// Error は error インターフェースを実装します。
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.original)
}

// Unwrap は元のエラーを返します。
func (e *AppError) Unwrap() error {
	return e.original
}

// NewAppError は新しい AppError を作成します。
func NewAppError(errType ErrorType, message string, original error) *AppError {
	return &AppError{
		Type:     errType,
		Message:  message,
		original: original,
	}
}
