package lib

type TransactionNotSupportedError struct{}

func (e *TransactionNotSupportedError) Error() string {
	return "transactions are only supported in the shell using semicolons to separate each statement.\nFor example: \"BEGIN; [your SQL statements]; END\""
}

type InvalidStatementsResult struct{}

func (e *InvalidStatementsResult) Error() string {
	return "invalid statements result"
}

type UnableToPrintStatementResult struct{}

func (e *UnableToPrintStatementResult) Error() string {
	return "unable to print statement result. You should check if its an error before printing it"
}
