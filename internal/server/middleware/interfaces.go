package middleware

type Logger interface {
	Infoln(args ...interface{})
}
