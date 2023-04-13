package db

type InvalidStatementsResult struct{}

func (e *InvalidStatementsResult) Error() string {
	return e.internalError()
}
func (e *InvalidStatementsResult) internalError() string {
	return "invalid statements result"
}

type UnableToPrintStatementResult struct{}

func (e *UnableToPrintStatementResult) Error() string {
	return e.internalError()
}
func (e *UnableToPrintStatementResult) internalError() string {
	return "unable to print statement result. You should check if its an error before printing it"
}
