package errors

// FailedHTTPReqError wraps error for http requests with code != 200
type FailedHTTPReqError struct {
	Err error
}

func (e FailedHTTPReqError) Error() string {
	return e.Err.Error()
}

// DatabaseError wraps gorm error
type DatabaseError struct {
	Err error
}

func (e DatabaseError) Error() string {
	return e.Err.Error()
}

// TelegramError wraps message sending error
type TelegramError struct {
	Err error
}

func (e TelegramError) Error() string {
	return e.Err.Error()
}
