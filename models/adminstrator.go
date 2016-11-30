package models

import (
	"errors"
	"net/http"
	"time"

	"github.com/jaybeecave/base/datastore"

	dat "gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"
	redis "gopkg.in/redis.v5"
)

// Pack Struct
type Administrator struct {
	AdministratorID int64        `db:"administrator_id" json:"PackID"`
	FirstName       string       `db:"first_name" json:"FirstName"`
	LastName        string       `db:"last_name" json:"LastName"`
	Email           string       `db:"email" json:"Email"`
	Password        string       `db:"password" json:"Password"`
	DateCreated     dat.NullTime `db:"date_created"  json:"DateCreated"`
	DateModified    dat.NullTime `db:"date_modified" json:"DateModified"`
}

// STRUCT HELPERS

// PackHelper provides a handle for Pack helper functions
type AdministratorHelper struct {
	DB    *runner.DB
	Cache *redis.Client
}

// Helper wraps the database connection and provides a handle for Pack helper functions
func NewAdministratorHelper(store *datastore.Datastore) *AdministratorHelper {
	helper := &AdministratorHelper{}
	helper.DB = store.DB
	helper.Cache = store.Cache
	return helper
}

func (h *AdministratorHelper) New() *Administrator {
	record := &Administrator{}
	// check DateCreated
	record.DateCreated = dat.NullTimeFrom(time.Now())
	return record
}

func (h *AdministratorHelper) NewFromRequest(req *http.Request) (*Administrator, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}

	record := h.New()
	err = h.UpdateFromRequest(req, record)
	if err != nil {
		return nil, err
	}

	// set DateCreated
	record.DateCreated = dat.NullTimeFrom(time.Now())
	record.DateModified = dat.NullTimeFrom(time.Now())

	return record, nil
}

func (h *AdministratorHelper) LoadAndUpdateFromRequest(req *http.Request) (*Administrator, error) {
	// dummyPack is used to get the pack ID from the request and also to check the date modified
	dummyRecord, err := h.NewFromRequest(req)
	if err != nil {
		return nil, err
	}

	if dummyRecord.AdministratorID <= 0 {
		return nil, errors.New("The pack failed to load because PackID was not found in the request.")
	}

	record, err := h.Load(dummyRecord.AdministratorID)
	if dummyRecord.DateModified.Valid && record.DateModified.Valid && dummyRecord.DateModified.Time.After(record.DateModified.Time) {
		return nil, errors.New("The pack failed to save because the DateModified value in the database is more recent then DateModified value on the request.")
	}

	err = h.UpdateFromRequest(req, record)

	// set DateCreated
	record.DateCreated = dat.NullTimeFrom(time.Now())

	return record, nil
}

func (h *AdministratorHelper) UpdateFromRequest(req *http.Request, record *Administrator) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}

	record.FirstName = req.PostFormValue("FirstName")
	record.LastName = req.PostFormValue("LastName")
	record.Email = req.PostFormValue("Email")
	record.Password = req.PostFormValue("Password")
	return nil
}

func (h *AdministratorHelper) Load(id int64) (*Administrator, error) {
	record := &Administrator{}
	err := h.DB.
		Select("administrator_id", "first_name", "last_name", "email", "password", "date_created", "date_modified").
		From("administrators").
		Where("user_id = $1", id).
		QueryStruct(record)

	if err != nil {
		return nil, err
	}

	return record, nil
}

func (h *AdministratorHelper) All() ([]*Administrator, error) {
	var records []*Administrator
	err := h.DB.Select("*").
		From("administrators").
		QueryStructs(&records)

	if err != nil {
		return nil, err
	}

	return records, nil
}

// Save a Uer returning the result
func (h *AdministratorHelper) Save(record *Administrator) error {

	// check DateCreated
	if !record.DateCreated.Valid {
		record.DateCreated = dat.NullTimeFrom(time.Now())
	}

	record.DateModified = dat.NullTimeFrom(time.Now())

	err := h.DB. // _ represents the Result struct which has number of records affected and lastinsertid both of which we dont really care about
			Upsert("administrators").
			Columns("first_name", "last_name", "email", "password", "date_created", "date_modified").
			Values(record.FirstName, record.LastName, record.Email, record.Password, record.DateCreated, record.DateModified).
			Where("administrator_id=$1", record.AdministratorID).
			Returning("administrator_id").
			QueryStruct(record)

	// log.Info(result)
	if err != nil {
		return err
	}

	return nil
}
