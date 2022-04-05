package util

import (
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
)

func InitSentry() {
	uri, err := url.Parse(os.Getenv("SENTRY"))
	if err != nil {
		logrus.Errorf("Failed to parse sentry URL: %v, integration disabled!", err)
		return
	}
	err = sentry.Init(sentry.ClientOptions{Dsn: uri.String()})
	if err != nil {
		logrus.Errorf("Failed to initialise sentry: %v", err)
	}
}
