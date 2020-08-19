package util

import (
	"net/http"
	"regexp"
)

var (
	CFConnectingIP = http.CanonicalHeaderKey("CF-Connecting-IP")
	XForwardedFor  = http.CanonicalHeaderKey("X-Forwarded-For")
	XRealIP        = http.CanonicalHeaderKey("X-Real-IP")
	Authorization  = http.CanonicalHeaderKey("Authorization")
	ContentType    = http.CanonicalHeaderKey("Content-Type")
	TokenPattern   = regexp.MustCompile("DELAPI_.{32}-([0-9]{17,20})")
)

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
