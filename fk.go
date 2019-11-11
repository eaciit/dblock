package dblock

import (
	"git.eaciitapp.com/sebar/dbflex"
	"github.com/eaciit/toolkit"
)

type FKField struct {
	Field    string
	FieldRef string
}

type FK struct {
	Table1 string
	Table2 string
	Fields []FKField
}

func NewFK(name1, name2 string, fields ...FKField) *FK {
	fk := new(FK)
	fk.Table1 = name1
	fk.Table2 = name2
	fk.Fields = fields
	return fk
}

func (fk *FK) WhereT1(data toolkit.M) *dbflex.Filter {
	ws := []*dbflex.Filter{}
	for _, f := range fk.Fields {
		ws = append(ws, dbflex.Eq(f.Field, data.Get(f.FieldRef)))
	}

	if len(ws) == 1 {
		return ws[0]
	} else if len(ws) > 1 {
		return dbflex.And(ws...)
	} else {
		return nil
	}
}

func (fk *FK) WhereT2(data toolkit.M) *dbflex.Filter {
	ws := []*dbflex.Filter{}
	for _, f := range fk.Fields {
		ws = append(ws, dbflex.Eq(f.FieldRef, data.Get(f.Field)))
	}

	if len(ws) == 1 {
		return ws[0]
	} else if len(ws) > 1 {
		return dbflex.And(ws...)
	} else {
		return nil
	}
}
