package dto

import (
	"errors"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
)

type LogData struct {
	TimeStamp time.Time    `csv:"timeStamp"`
	NickName  string       `csv:"nickName"`
	Action    enums.Action `csv:"action"`
	IPAddress string       `csv:"ipAddress"`
	Country   string       `csv:"country"`
}

func (l *LogData) Validate() error {
	if l.TimeStamp.IsZero() {
		return errors.New("invalid timestamp")
	}
	if l.NickName == "" {
		return errors.New("invalid nickname")
	}
	if !l.Action.IsValid() {
		return errors.New("invalid action")
	}
	return nil
}
