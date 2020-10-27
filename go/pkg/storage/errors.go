package storage

type DoesNotExistError struct {
	msg string
}

func (e *DoesNotExistError) Error() string {
	return e.msg
}
