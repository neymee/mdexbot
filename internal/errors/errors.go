package errors

// FailedHTTPReqError wraps error for http requests with code != 200
type FailedHTTPReqError struct {
	Err error
}

func (e FailedHTTPReqError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "Unknown http request error"
}

// DatabaseError wraps gorm error
type DatabaseError struct {
	Err error
}

func (e DatabaseError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "Unknown database error"
}

// TelegramError wraps message sending error
type TelegramError struct {
	Err error
}

func (e TelegramError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "Unknown telegram error"
}
