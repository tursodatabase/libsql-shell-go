package enums

type PrintMode string

const (
	TABLE_MODE PrintMode = "table"
	CSV_MODE   PrintMode = "csv"
	JSON_MODE  PrintMode = "json"
)

type HistoryMode int

const (
	SingleHistory HistoryMode = iota
	PerDatabaseHistory
	LocalHistory
)
