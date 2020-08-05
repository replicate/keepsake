package storage

type NotExistError struct {
	msg string
}

func (e *NotExistError) Error() string {
	return e.msg
}
