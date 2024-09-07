package core

const (
	DATABASE    = "database"
	MIDDLLEWARE = "middlleware"
)

type Service interface {
	ExecuteOnce()
	AutoGather()
	Shell()
	SetHost(string)
	SetPort(int)
	SetCmd(string)
}
