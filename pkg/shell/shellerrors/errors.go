package shellerrors

type InternalError interface {
	error
	internalError() string
}

type UserError interface {
	error
	userError() string
}

type TransactionNotSupportedError struct{}

func (e *TransactionNotSupportedError) Error() string {
	return e.userError()
}
func (e *TransactionNotSupportedError) userError() string {
	return "transactions are only supported in the shell using semicolons to separate each statement.\nFor example: \"BEGIN; [your SQL statements]; END\""
}

type UrlDoesNotContainUserError struct{}

func (e *UrlDoesNotContainUserError) Error() string {
	return e.userError()
}
func (e *UrlDoesNotContainUserError) userError() string {
	return "url does not contain user"
}

type InvalidTursoProtocolError struct{}

func (e *InvalidTursoProtocolError) Error() string {
	return e.userError()
}
func (e *InvalidTursoProtocolError) userError() string {
	return "invalid turso protocol. valid protocols are libsql://, wss:// and ws://"
}
