package global

import "database/sql"

const (
	DataFolder = "data/"
	UrlFile    = DataFolder + "url.txt"
	DateFormat = "2006-01-02"
)

var (
	DB                 *sql.DB
	UseDatabase        bool
	SendToTelegramFlag bool
	ProxyURL           string
	Help               bool
)
