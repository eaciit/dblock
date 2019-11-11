package dblock_test

import (
	"fmt"
	"testing"

	_ "github.com/eaciit/flexmgo"

	"git.eaciitapp.com/sebar/dbflex"

	"github.com/eaciit/dblock"
	"github.com/eaciit/toolkit"
	"github.com/smartystreets/goconvey/convey"
)

var (
	connStringDefault = "mongodb://localhost:27017/dblock"
)

func TestMain(t *testing.T) {
	convey.Convey("Prep dep", t, func() {
		defer func() {
			conn, _ := dbflex.NewConnectionFromURI(connStringDefault, nil)
			conn.Connect()
			conn.Execute(dbflex.From("countries").Delete(), nil)
			conn.Execute(dbflex.From("emptable").Delete(), nil)
			conn.Close()
		}()

		deps := []*dblock.FK{
			//dblock.NewFK("emprelations", "emptable", dblock.FKField{"empid", "_id"}),
			dblock.NewFK("emptable", "countries", dblock.FKField{"countryid", "_id"}),
		}

		lock := dblock.NewLock(func() (dbflex.IConnection, error) {
			if c, e := dbflex.NewConnectionFromURI(connStringDefault, nil); e != nil {
				return nil, fmt.Errorf("unable to create conn. %s", e.Error())
			} else {
				if e = c.Connect(); e != nil {
					return nil, fmt.Errorf("unable to connect to connection. %s", e.Error())
				}
				return c, nil
			}
		}).SetDeps(deps)

		convey.Convey("insert country data", func() {
			err := lock.Save("countries",
				toolkit.M{}.Set("_id", "ID").Set("name", "Indonesia"),
				toolkit.M{}.Set("_id", "SG").Set("name", "Singapore"))
			convey.So(err, convey.ShouldBeNil)

			convey.Convey("insert empl", func() {
				err := lock.Save("emptable",
					toolkit.M{}.Set("_id", "E1").Set("name", "Employee1").Set("countryid", "ID"))
				convey.So(err, convey.ShouldBeNil)

				err = lock.Save("emptable",
					toolkit.M{}.Set("_id", "E2").Set("name", "Employee2").Set("countryid", "MY"))
				convey.So(err, convey.ShouldNotBeNil)

				convey.Convey("delete relevant table", func() {
					err := lock.Delete("countries", dbflex.Eq("_id", "ID"))
					convey.So(err, convey.ShouldNotBeNil)

					err = lock.Delete("countries", dbflex.Eq("_id", "SG"))
					convey.So(err, convey.ShouldBeNil)
				})

				convey.Convey("Validation", func() {
					err := lock.ValidateChild("countries", toolkit.M{}.Set("_id", "ID"), dblock.Exists)
					convey.So(err, convey.ShouldBeNil)

					err = lock.ValidateChild("countries", toolkit.M{}.Set("_id", "ID"), dblock.NotExists)
					convey.So(err, convey.ShouldNotBeNil)

					err = lock.ValidateParent("emptable", toolkit.M{}.Set("_id", "E3").Set("countryid", "SG"), dblock.Exists)
					convey.So(err, convey.ShouldBeNil)

					err = lock.ValidateParent("emptable", toolkit.M{}.Set("_id", "E3").Set("countryid", "NZ"), dblock.NotExists)
					convey.So(err, convey.ShouldBeNil)
				})
			})
		})
	})
}
