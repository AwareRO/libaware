package middlewares

import (
	"github.com/julienschmidt/httprouter"
)

type Wrapper interface {
	Wrap(nextHandler httprouter.Handle) httprouter.Handle
}

type MetricsWrapper interface {
	Wrapper
	MetricsHandler() httprouter.Handle
}
