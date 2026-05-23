package jobs

import (
	"context"
	"time"
)

type sessionExpirationStore interface {
	ExpireSessions(context.Context, time.Time) (int, error)
}

func ExpireSessions(ctx context.Context, repo any, now time.Time) (int, error) {
	expirer, ok := repo.(sessionExpirationStore)
	if !ok {
		return 0, nil
	}
	return expirer.ExpireSessions(ctx, now)
}
