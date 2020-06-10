package database

type RecordNotFoundError struct {
	s string
}

func (r RecordNotFoundError) Error() string {
	return r.s
}

func NewRecordNotFoundError() RecordNotFoundError {
	return RecordNotFoundError{s: "Record not found in database"}
}
