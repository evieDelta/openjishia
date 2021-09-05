package enpsql

import (
	"embed"
	"fmt"
	"io/fs"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/eviedelta/openjishia/internal/stopwatch"
	"github.com/eviedelta/openjishia/wlog"
	"github.com/gocraft/dbr/v2"
	"github.com/pkg/errors"
)

// RegisterSchema imports a list of schema updates to import from.
// keeps track of the schema versions currently in the database and will update it as new schemas updated are added.
// as mentioned in the package doc each package should also create its own SQL schema and avoid using the public schema or modifying schemas other than its own
//
// Note that there is no gaurenteed order in which package schemas are updated
// an update in package A may be applied before or after package B even if they were pushed at the same time or have the same version number
// as such a schema update should not rely on another packages schema having recieved any particular schema update even if they are distributed together
// You may only rely on schema versions within a package being in the correct order, not the order of other packages
// (Eg, pkgA v3 will always happen before pkgA v4, but pkgB v3 may happen at any time relative to pkgA v3)
//
// Schema versions should not be changed once they are in use, if a change is needed a new schema update should be added instead
//
// the name should be treated as effectively static once in use,
// while theoretically possible to migrate doing such should be avoided if at all possible
func RegisterSchema(name string, updates []string) {
	schemaList[name] = updates
}

// RegisterSchemaFS is like RegisterSchema but it loads from a fs.FS (see RegisterSchema for more info).
// it expects the files to be formatted in the form of a decimal number followed by the .sql extension
// eg, 001.sql, 002.sql ... 015.sql ...
// They should be numbered starting from 1, and no numbers should be skipped, no files other than numbered sql files should be contained in the folder
// dir refers to the subdirectory in the fs.FS to load, as to make it easier to use things like embed.FS
func RegisterSchemaFS(name string, f fs.FS, dir string) error {
	d, err := fs.Glob(f, path.Join(dir, "*.sql"))
	if err != nil {
		return err
	}

	nd := make([]string, len(d))

	isNum := func(s string) bool {
		for _, x := range s {
			if !unicode.IsDigit(x) {
				return false
			}
		}
		return true
	}

	for _, x := range d {
		y := strings.TrimSuffix(x, ".sql")
		y = strings.Split(y, "/")[strings.Count(y, "/")]

		if !isNum(y) {
			continue
		}
		d, err := fs.ReadFile(f, x)
		if err != nil {
			return fmt.Errorf("%v: %w", x, err)
		}

		i, err := strconv.Atoi(y)
		if err != nil {
			return fmt.Errorf("%v: %w", x, err)
		}

		nd[i-1] = string(d)
	}

	RegisterSchema(name, nd)
	return nil
}

var schemaList = make(map[string][]string)

type schema struct {
	Name    string `db:"name"`
	Version int    `db:"version"`
}

var schemas = struct {
	Name    string
	Version string
}{
	Name:    "name",
	Version: "version",
}

const schemaKeyInternal = "~enpsql:internal/public"

//go:embed schema/*
var internalSchema embed.FS

func init() {
	err := RegisterSchemaFS(schemaKeyInternal, internalSchema, "schema")
	if err != nil {
		panic(err)
	}
}

// the db backup for schema updates only needs to happen once, so we make sure it doesn't
var backupOnce sync.Once

func doBackup() (do bool) {
	// no backup set
	if Config.External.BackupCommand == "" {
		return true
	}
	do = true

	// do command
	backupOnce.Do(func() {
		c := exec.Command("sh", "-c", Config.External.BackupCommand)
		o, err := c.CombinedOutput()
		if err != nil {
			do = false // don't
			wlog.Info.Print("error running backup command!", err)
		}
		wlog.Info.Print("backup output:\n", string(o))
	})
	return do
}

// goes through and checks for schema updates and executes any new schema updates if necessary
func updateAll() error {
	for i, x := range schemaList {
		if i == schemaKeyInternal {
			// this one is handled separately so we need not worry here
			continue
		}

		// any update needed?
		td, cr, nw, err := schemaUpdateChecker(i, x)
		if err != nil {
			return err
		}

		// no? then skip
		if len(td) == 0 {
			continue
		}

		// do the backup thingy
		ok := doBackup()
		if !ok {
			return errors.New("backup suspected to have failed")
		}

		// apply the updates (hopefully)
		err = schemaApplyUpdates(i, td, cr, nw)
		if err != nil {
			return err
		}
	}

	return nil
}

func initialCheck() error {
	var s bool

	// does the schema table exist?
	err := ses.SelectBySql(`SELECT EXISTS(SELECT * FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'schemas')`).
		LoadOne(&s)
	if err != nil {
		return errors.Wrap(err, "Check Exists")
	}

	if !s {
		// if it doesn't then we need to apply v1
		err := schemaApplyUpdates(schemaKeyInternal, schemaList[schemaKeyInternal], 0, true)
		return errors.Wrap(err, "Init Database")
	}

	// check for updates to the root schema
	td, cr, nw, err := schemaUpdateChecker(schemaKeyInternal, schemaList[schemaKeyInternal])
	if err != nil {
		return errors.Wrap(err, "Check updates on root schema")
	}
	if len(td) == 0 {
		// if there are none we are done
		return nil
	}

	// do backup thingy if we have any updates to apply
	ok := doBackup()
	if !ok {
		return errors.New("backup suspected to have failed")
	}

	// apply updates
	arr := schemaApplyUpdates(schemaKeyInternal, td, cr, nw)
	return errors.Wrap(arr, "Update root schema")
}

// schemaUpdateChecker takes the list of schema patches and compares it to the current version number in the database
// if there are updates needed it returns the updates that need to be done, the current version, and if this schema is not currently present in the database
func schemaUpdateChecker(name string, versions []string) (todo []string, current int, new bool, err error) {
	s, err := schemaGet(name)
	if err != nil {
		return nil, 0, false, err
	}
	// version 0 means it doesn't exist yet / is new
	new = s.Version == 0

	current = s.Version
	todo = versions[current:]

	return todo, current, new, nil
}

// schemaApplyUpdates applies a list of schema updates, taking pretty much the output of schemaUpdateChecker
func schemaApplyUpdates(name string, updates []string, current int, new bool) error {
	// we do this on a transaction so we can rollback if it fails and won't be left manually fixing half done limbo
	tx, err := ses.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	// keep a log of the schema update events
	// good to be clear with this
	st := newLogBlock()
	defer wlog.Info.Print(st)

	st.Printf("Applying migrations for `%v`. `%v` migrations to apply", name, len(updates))

	// timer to keep an eye on how long it takes, cause mmm tasty stats
	timer := stopwatch.NewAndStart()

	// go through all updates
	for i, x := range updates {
		// execute the sql code
		r, err := tx.Exec(x)
		if err != nil {
			st.Printf("- Error applying v`%v` reverting and erroring out: %v", current+i+1, err)
			return err
		}

		// mmm stats, this time it tells how many rows were changed
		raf, err := r.RowsAffected()
		switch {
		case err != nil:
			st.Printf("- Error getting rows affected: `%v`", err)
			fallthrough
		case raf == 0:
			st.Printf("- Applied migration v`%v`",
				current+i+1)
		case raf != 0:
			st.Printf("- Applied migration v`%v`, Rows Affected: `%v`",
				current+i+1, raf)
		}
	}
	st.Printf("Finished writing migrations")

	// now we update the meta key
	if new {
		// if its new we insert a new meta key
		// (we have to write this again and can't use the schemaNew() because we are using a transaction)
		_, err = tx.InsertInto(table.Schemas).
			Columns(schemas.Name, schemas.Version).
			Values(name, len(updates)).
			Exec()
		if err != nil {
			st.Printf("- Error creating version tag, reverting and erroring out: %v", err)
			return err
		}
	} else {
		// if it already exists, we can then update the existing one
		_, err = tx.Update(table.Schemas).
			Where(dbr.Eq(schemas.Name, name)).
			IncrBy(schemas.Version, len(updates)). // increment the version key by the number of updates done
			Exec()
		if err != nil {
			st.Printf("- Error updating version tag, reverting and erroring out: %v", err)
			return err
		}
	}

	st.Printf("Finished updating meta, commiting")

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		st.Printf("- Error commiting transaction, reverting and erroring out: %v", err)
		return err
	}

	st.Printf("Finished migration for `%v`, v`%v`>`%v`, took `%v`",
		name, current, current+len(updates), timer.Stop())

	return nil
}

// schemaGet obtains the meta key for a schema name
func schemaGet(name string) (s schema, err error) {
	err = ses.Select("*").
		From(table.Schemas).
		Where(dbr.Eq(schemas.Name, name)).
		LoadOne(&s)

	if errors.Is(err, dbr.ErrNotFound) {
		return schema{
			Name:    name,
			Version: 0,
		}, nil
	}

	return s, err
}

/*
// schemaUpdate updates the value of a schema meta
// i don't think this is actually used at all
func schemaUpdate(name string, version int) error {
	_, err := ses.Update(table.Schemas).
		Where(dbr.Eq(schemas.Name, name)).
		Set(schemas.Version, version).
		Exec()

	return err
}

// schemaNew creates a new schema meta entry
// i don't think this is actually used at all
func schemaNew(name string, version int) error {
	_, err := ses.InsertInto(table.Schemas).
		Columns(schemas.Name, schemas.Version).
		Values(name, version).
		Exec()

	return err
}
*/
