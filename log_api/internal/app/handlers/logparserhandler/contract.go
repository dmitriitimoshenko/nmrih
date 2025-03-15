package logparserhandler

import "time"

type Service interface {
	Parse(requestTimeStamp time.Time) error
}
