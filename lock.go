package dblock

import (
	"github.com/ariefdarmawan/sharedobject"

	"git.eaciitapp.com/sebar/dbflex"
)

type (
	ExistsNotExists string
)

const (
	Exists    ExistsNotExists = "Exists"
	NotExists ExistsNotExists = "NotExists"
)

/*
type DBLock interface {
	Execute(dbflex.ICommand, toolkit.M) error
	Cursor(dbflex.ICommand, toolkit.M) (dbflex.ICursor, error)
	SetDeps([]*FK)
	ValidateChild(string, interface{}, ExistsNotExists)
	ValidateParent(string, interface{}, ExistsNotExists)
}
*/

func NewLock(conn func() (dbflex.IConnection, error)) *FlexLock {
	l := new(FlexLock)
	l.connFn = conn
	l.locks = sharedobject.NewSharedData()
	return l
}
