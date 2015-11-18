package migration

import (
	"database/sql"
	"fmt"
	"os"
	// "strconv"
	"strings"

	"github.com/codegangsta/cli"
)

type Migration interface {
	Up() error
	Down() error
	Name() string
	Version() string
}

var migrations []Migration
var db *sql.DB

func Run(ms []Migration, _db *sql.DB) {
	db = _db
	migrations = ms
	app := cli.NewApp()
	app.Name = "migration"
	app.Usage = "migration <command> <options> <args>"
	app.Commands = []cli.Command{
		{
			Name:   "status",
			Usage:  "migration status",
			Action: status,
		},
		{
			Name:   "reset",
			Usage:  "migration reset",
			Action: reset,
		},
		{
			Name:   "up",
			Usage:  "migration up [<version>]",
			Action: up,
		},
		{
			Name:   "down",
			Usage:  "migration down [<version>]",
			Action: down,
		},
	}
	app.Run(os.Args)
	return
}

func status(c *cli.Context) {
	sms := loadSchemeMigrations(migrations, db)
	nameLen := 6
	versionLen := 9
	appliedAtLen := 21
	for _, sm := range sms {
		if sm.Migration == nil {
			break
		}
		vLen := len(sm.Migration.Version()) + 2
		nLen := len(sm.Migration.Name()) + 2
		if vLen > versionLen {
			versionLen = vLen
		}
		if nLen > nameLen {
			nameLen = nLen
		}
	}
	format := fmt.Sprintf("%%-%ds%%-%ds%%-%ds\n", versionLen, nameLen, appliedAtLen)
	fmt.Println("")
	fmt.Printf(format, "Version", "Name", "AppliedAt")
	fmt.Println(strings.Repeat("-", nameLen+versionLen+appliedAtLen+1))
	for _, sm := range sms {
		if sm.Migration == nil {
			break
		}
		fmt.Printf(
			format,
			sm.Migration.Version(),
			sm.Migration.Name(),
			sm.appliedAtStr(),
		)
	}
	fmt.Println("")
}

func reset(c *cli.Context) {

}

func up(c *cli.Context) {
	tx, err := db.Begin()
	if err != nil {
		panic(err.Error())
	}

	sms := loadSchemeMigrations(migrations, db)
	for _, sm := range sms {
		if !sm.appliedAt.IsZero() {
			continue
		}
		if err := sm.Migration.Up(); err != nil {
			txerr := tx.Rollback()
			if txerr != nil {
				fmt.Printf(txerr.Error())
				panic(err.Error())
			}
			panic(err.Error())
		}
		if err := sm.save(db); err != nil {
			txerr := tx.Rollback()
			if txerr != nil {
				fmt.Printf(txerr.Error())
				panic(err.Error())
			}
			panic(err.Error())
		}
	}
	if err := tx.Commit(); err != nil {
		panic(err.Error())
	}
}

func down(c *cli.Context) {
	tx, err := db.Begin()
	if err != nil {
		panic(err.Error())
	}
	sms := loadSchemeMigrations(migrations, db)
	for i := len(sms) - 1; i >= 0; i-- {
		sm := sms[i]
		if sm.appliedAt.IsZero() {
			continue
		}
		if err := sm.Migration.Down(); err != nil {
			txerr := tx.Rollback()
			if txerr != nil {
				fmt.Printf(txerr.Error())
				panic(err.Error())
			}
			panic(err.Error())
		}
		if err := sm.destroy(db); err != nil {
			txerr := tx.Rollback()
			if txerr != nil {
				fmt.Printf(txerr.Error())
				panic(err.Error())
			}
			panic(err.Error())
		}
	}
	if err := tx.Commit(); err != nil {
		panic(err.Error())
	}
}
