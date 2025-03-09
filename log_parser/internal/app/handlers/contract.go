package handlers

import "time"

type Service interface {
	Parse(requestTimeStamp time.Time) error
}
