package util

import (
	"context"
	"github.com/discordextremelist/api/database"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	Database = database.NewManager()
	Router   chi.Router
	Dev      bool
)

func Scan[T any](key string) map[string]T {
	m := make(map[string]T)
	var cursor uint64 = 0
	for {
		res := Database.Redis.HScan(context.TODO(), key, cursor, "", 0)
		if res.Err() != nil {
			sentry.CaptureException(res.Err())
			logrus.Errorf("Scan failed: %v", res.Err())
			continue
		}
		keys, c, _ := res.Result()
		l := len(keys)
		for i := 0; i < l; i += 2 {
			end := i + 2
			if end > l {
				end = l
			}
			v := keys[i:end]
			var val T
			err := json.Unmarshal([]byte(v[1]), &val)
			if err != nil {
				continue
			}
			m[v[0]] = val
		}
		cursor = c
		if c == 0 {
			break
		}
	}
	return m
}
