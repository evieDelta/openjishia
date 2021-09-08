package nschedule

import (
	"time"

	"github.com/eviedelta/openjishia/nschedule/insched"
	"github.com/gocraft/dbr/v2"
)

var table = struct {
	Events string
}{
	Events: "nscheduler.events",
}

var events = struct {
	ID         string
	Action     string
	Details    string
	DeferCount string
	Time       string
}{
	ID:         "id",
	Action:     "action",
	Details:    "details",
	DeferCount: "defer_count",
	Time:       "time",
}

type Database struct {
	ss *dbr.Session
}

func (db *Database) Add(e insched.Entry) (insched.Entry, error) {
	//e.Time = e.Time.UTC()
	err := db.ss.InsertInto(table.Events).Columns(
		events.Action,
		events.Details,
		events.DeferCount,
		events.Time,
	).Record(&e).
		Returning(events.ID).
		Load(&e.ID)

	if err != nil {
		return e, err
	}

	return db.Get(e.ID)
}

func (db *Database) Get(id uint64) (e insched.Entry, err error) {
	err = db.ss.Select("*").
		From(table.Events).
		Where(dbr.Eq(events.ID, id)).
		LoadOne(&e)

	return e, err
}

func (db *Database) GetPending(action string, t time.Time) (el []insched.Entry, err error) {
	_, err = db.ss.Select("*").
		From(table.Events).
		Where(dbr.Eq(events.Action, action)).
		Where(dbr.Lte(events.Time, t)).
		Load(&el)

	return el, err
}

func (db *Database) Remove(id uint64) (err error) {
	_, err = db.ss.DeleteFrom(table.Events).
		Where(dbr.Eq(events.ID, id)).
		Exec()

	return err
}

func (db *Database) Defer(id uint64, by time.Duration) (err error) {
	_, err = db.ss.Update(table.Events).
		Where(dbr.Eq(events.ID, id)).
		IncrBy(events.Time, by). // hoping incrby works here
		Exec()

	return err
}

func (db *Database) Reschedule(id uint64, to time.Time) (err error) {
	_, err = db.ss.Update(table.Events).
		Where(dbr.Eq(events.ID, id)).
		Set(events.Time, to).
		Exec()

	return err
}
