package global

const (
	DataFolder = "data/"
	UrlFile    = DataFolder + "url.txt"
	DateFormat = "2006-01-02"
)

var (
	UseDatabase        bool
	SendToTelegramFlag bool
	ProxyURL           string
	Help               bool
)
