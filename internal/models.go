package models

type Server struct {
	ID   int
	FQDN string
	User string
	Pass string
}

type ServerLog struct {
	ID       int
	ServerID int
	Log      string
}
