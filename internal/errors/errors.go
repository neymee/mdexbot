package errors

import "errors"

var (
	// FailedHTTPReqError wraps error for http requests with code != 200
	FailedHTTPReqError = errors.New("failed http request")
	// DatabaseError wraps gorm error
	DatabaseError = errors.New("database error")
	// TelegramError wraps message sending error
	TelegramError = errors.New("telegram error")
)
