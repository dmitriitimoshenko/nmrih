package logparserhandler

import "time"

type service interface {
	Parse(requestTimeStamp time.Time) error
}
