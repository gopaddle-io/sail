package trace

import (
	"github.com/sirupsen/logrus"
)

type SailContext struct {
	BaseContext
}

func NewSailContext(log *logrus.Entry, requestID string) *SailContext {
	sailCxt := SailContext{}
	sailCxt.BaseContext.RequestID = requestID
	sailCxt.BaseContext.Log = log
	// sailCxt.BaseContext.AccountID = accountID
	return &sailCxt
}

type BaseContext struct {
	// AccountID string
	RequestID string
	InternalContext
}

type InternalContext struct {
	Log *logrus.Entry
}
