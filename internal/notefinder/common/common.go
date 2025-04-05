package common

const (
	AppName    = "Notefinder"
	AppVersion = 0.1
	ConfigPath = ".config/notefinder.ini"
)

type Request int

const (
	RequestLoadData Request = iota
	RequestStop
)
