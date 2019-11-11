package dblock

import (
	"fmt"
	"sync"

	"github.com/ariefdarmawan/sharedobject"
	"github.com/eaciit/toolkit"

	"git.eaciitapp.com/sebar/dbflex"
)

type FlexLock struct {
	connFn func() (dbflex.IConnection, error)
	//lock   *sync.RWMutex

	deps  []*FK
	locks *sharedobject.SharedData
}

func (l *FlexLock) SetDeps(deps []*FK) *FlexLock {
	l.deps = deps
	return l
}

func (l *FlexLock) getLock(name string) *sync.RWMutex {
	if l.locks == nil {
		l.locks = sharedobject.NewSharedData()
	}
	if lock := l.locks.Get(name); lock == nil {
		lockMutex := new(sync.RWMutex)
		l.locks.Set(name, lockMutex)
		return lockMutex
	} else {
		return lock.(*sync.RWMutex)
	}
}

func (l *FlexLock) ValidateChild(name string, data interface{}, en ExistsNotExists) error {
	mdata, err := toolkit.ToM(data)
	if err != nil {
		return fmt.Errorf("unable to serialize data. %s", err.Error())
	}

	relevantDeps := []*FK{}
	for _, dep := range l.deps {
		if dep.Table2 == name {
			relevantDeps = append(relevantDeps, dep)
		}
	}

	for _, dep := range relevantDeps {
		depError := func() error {
			conn, err := l.connFn()
			if err != nil {
				return fmt.Errorf("unable to connect. %s", err.Error())
			}
			defer conn.Close()
			filter := dep.WhereT1(mdata)
			query := dbflex.From(dep.Table1).Where(filter).Select().Take(1)
			cursor := conn.Cursor(query, nil)
			if cursor.Error() != nil {
				return fmt.Errorf("unable to create cursor. %s", err.Error())
			}
			ms := []toolkit.M{}
			if err = cursor.Fetchs(&ms, 0); err != nil {
				return fmt.Errorf("unable to fetch cursor. %s", err.Error())
			}
			if len(ms) == 0 {
				if en != NotExists {
					return fmt.Errorf("data exists on %s", dep.Table1)
				}
			} else {
				if en != Exists {
					return fmt.Errorf("data not exists on %s", dep.Table1)
				}
			}

			return nil
		}()
		if depError != nil {
			return depError
		}
	}

	return nil
}

func (l *FlexLock) ValidateParent(name string, data interface{}, en ExistsNotExists) error {
	mdata, err := toolkit.ToM(data)
	if err != nil {
		return fmt.Errorf("unable to serialize data. %s", err.Error())
	}

	relevantDeps := []*FK{}
	for _, dep := range l.deps {
		if dep.Table1 == name {
			relevantDeps = append(relevantDeps, dep)
		}
	}

	for _, dep := range relevantDeps {
		depError := func() error {
			conn, err := l.connFn()
			if err != nil {
				return fmt.Errorf("unable to connect. %s", err.Error())
			}
			defer conn.Close()

			filter := dep.WhereT2(mdata)
			query := dbflex.From(dep.Table2).Where(filter).Select().Take(1)
			cursor := conn.Cursor(query, nil)
			if cursor.Error() != nil {
				return fmt.Errorf("unable to create cursor. %s: %s", dep.Table2, err.Error())
			}
			ms := []toolkit.M{}
			if err = cursor.Fetchs(&ms, 0); err != nil {
				if err.Error() != "EOF" {
					return fmt.Errorf("unable to fetch cursor. %s: %s", dep.Table2, err.Error())
				}
			}
			if en == Exists && len(ms) == 0 {
				return fmt.Errorf("data is not exists on %s with filter %s", dep.Table2, toolkit.JsonString(filter))
			} else if en == NotExists && len(ms) > 0 {
				return fmt.Errorf("data is exists on %s with filter %s", dep.Table2, toolkit.JsonString(filter))
			}
			return nil
		}()
		if depError != nil {
			return depError
		}
	}

	return nil
}

func (l *FlexLock) Save(name string, data ...interface{}) error {
	for _, d := range data {
		dErr := func() error {
			err := l.ValidateParent(name, d, Exists)
			if err != nil {
				return fmt.Errorf("Data consistency error. %s", err.Error())
			}

			querySave := dbflex.From(name).Save()
			connSave, err := l.connFn()
			if err != nil {
				return fmt.Errorf("unable to save. connection fail. %s", err.Error())
			}
			defer connSave.Close()

			if _, err = connSave.Execute(querySave, toolkit.M{}.Set("data", d)); err != nil {
				return fmt.Errorf("unable to save. connection fail. %s", err.Error())
			}
			return nil
		}()
		if dErr != nil {
			return dErr
		}
	}
	return nil
}

func (l *FlexLock) Delete(name string, where *dbflex.Filter) error {
	conn, err := l.connFn()
	if err != nil {
		return fmt.Errorf("unable to delete. connection fail. %s", err.Error())
	}
	defer conn.Close()

	//-- get items to be deleted
	cursor := conn.Cursor(dbflex.From(name).Where(where), nil)
	if cursor.Error() != nil {
		return fmt.Errorf("unable to get cursor for deleted items. %s", err.Error())
	}

	ms := []toolkit.M{}
	if err = cursor.Fetchs(&ms, 0); err != nil && err.Error() != "EOF" {
		return fmt.Errorf("unable to fetch cursor of deleted items. %s", err.Error())
	}

	for _, m := range ms {
		err := l.ValidateChild(name, m, NotExists)
		if err != nil {
			return fmt.Errorf("data consistency error. %s", err.Error())
		}
	}

	query := dbflex.From(name).Where(where).Delete()
	if _, err = conn.Execute(query, nil); err != nil {
		return fmt.Errorf("unable to delete. connection fail. %s", err.Error())
	}
	return nil
}

/*
func (l *FlexLock) Execute(cmd dbflex.ICommand, parm toolkit.M) error {
	if l.connFn == nil {
		return fmt.Errorf("connection is not yet implemented")
	}

	conn, err := l.connFn()
	if err != nil {
		return fmt.Errorf("connection error. %s", err.Error())
	}
	defer conn.Close()

	l.lock.Lock()
	_, err = conn.Execute(cmd, parm)
	l.lock.Unlock()
	return err
}

func (l *FlexLock) Cursor(cmd dbflex.ICommand, parm toolkit.M) (dbflex.ICursor, error) {
	if l.connFn == nil {
		return nil, fmt.Errorf("connection is not yet implemented")
	}

	conn, err := l.connFn()
	if err != nil {
		return nil, fmt.Errorf("connection error. %s", err.Error())
	}
	defer conn.Close()

	l.lock.RLock()
	cur := conn.Cursor(cmd, parm)
	err = cur.Error()
	if err != nil {
		return nil, err
	}
	l.lock.Unlock()

	return cur, nil
}
*/
