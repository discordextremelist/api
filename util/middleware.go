package util

import (
	"fmt"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var (
	CFConnectingIP = http.CanonicalHeaderKey("CF-Connecting-IP")
	XForwardedFor  = http.CanonicalHeaderKey("X-Forwarded-For")
	XRealIP        = http.CanonicalHeaderKey("X-Real-IP")
	Authorization  = http.CanonicalHeaderKey("Authorization")
)

func TokenValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Dev {
			auth := r.Header.Get(Authorization)
			if auth == "" {
				WriteJson(http.StatusForbidden, w, NoAuthError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RealIP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfConnecting := r.Header.Get(CFConnectingIP); cfConnecting != "" {
			r.RemoteAddr = cfConnecting
		} else if xff := r.Header.Get(XForwardedFor); xff != "" {
			r.RemoteAddr = xff
		} else if xri := r.Header.Get(XRealIP); xri != "" {
			r.RemoteAddr = xri
		} else {
			r.RemoteAddr = "127.0.0.1"
		}
		handler.ServeHTTP(w, r)
	})
}

func doLog(start time.Time, w middleware.WrapResponseWriter, r *http.Request) {
	log.Info(fmt.Sprintf(
		`%s - "%s %s %s" %d %d %s`,
		r.RemoteAddr,
		r.Method,
		r.URL,
		r.Proto,
		w.BytesWritten(),
		w.Status(),
		time.Since(start),
	))
}

func RequestLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer doLog(start, ww, r)
		handler.ServeHTTP(ww, r)
	})
}

func NotFound(w http.ResponseWriter, _ *http.Request) {
	WriteJson(http.StatusNotFound, w, NotFoundError)
}