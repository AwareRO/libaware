package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AwareRO/libaware/golang/http/identity"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

type ExtraField func(r *http.Request, params httprouter.Params) (string, string)

func ipv4(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	return ip
}

func isCrawler(r *http.Request) string {
	return fmt.Sprintf("%v", identity.NewMileusna()(r.UserAgent()))
}

func LogRequest(nextHandler httprouter.Handle, extras ...ExtraField) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		logger := log.Info().
			Str("endpoint", r.URL.String()).
			Str("method", r.Method).
			Str("ip", ipv4(r)).
			Str("crawler", isCrawler(r))
		for _, extra := range extras {
			logger = logger.Str(extra(r, params))
		}
		logger.Msg("Got request")
		nextHandler(w, r, params)
	}
}
