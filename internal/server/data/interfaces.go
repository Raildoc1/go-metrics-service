package data

type Logger interface {
	Infoln(args ...interface{})
	Debugln(args ...interface{})
	Errorln(args ...interface{})
}
