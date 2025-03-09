package models

import (
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/enums"
)

type LogData struct {
	TimeStamp time.Time    `csv:"timeStamp"`
	NickName  string       `csv:"nickName"`
	Action    enums.Action `csv:"action"`
	IPAddress string       `csv:"ipAddress"`
}
