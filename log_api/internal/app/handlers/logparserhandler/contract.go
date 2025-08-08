package logparserhandler

import (
	"context"
	"time"
)

type redisCache interface {
	FlushAll(ctx context.Context) error
}

type service interface {
	Parse(requestTimeStamp time.Time) error
}
