package logger

import "go.uber.org/zap"

// Filed types for structured logging. Most fileds are lazily marshaled,
// so it is inexpensive to add fileds to disabled log statements.
var (
	Any       = zap.Any
	Int       = zap.Int
	Namespace = zap.Namespace
)
