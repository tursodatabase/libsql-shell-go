package lib

type TransactionNotSupportedError struct{}

func (e *TransactionNotSupportedError) Error() string {
	return "transactions are only supported in the shell using semicolons to separate each statement.\nFor example: \"BEGIN; [your SQL statements]; END\""
}
