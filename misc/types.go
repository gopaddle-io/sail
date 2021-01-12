package misc

import (
	"github.com/sirupsen/logrus"
)

type InternalContext struct {
	// DS   dao.ReadWritter
	Log *logrus.Entry
	// HTTP bhttp.HTTPInterface
}

type MiscHandle struct {
	InternalContext
	AccountID string
	RequestID string
}

type BaseContext struct {
	AccountID string
	RequestID string
	InternalContext
	MiscHandle MiscHandle
}

type Response struct {
	Code     int
	Response string
}

type Error struct {
	Message string `json:"message"`

	RequestID string `json:"requestID"`
}
