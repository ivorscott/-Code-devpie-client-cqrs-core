package schema

import (
	"fmt"
	"github.com/devpies/devpie-client-core/users/platform/database"
	"io/ioutil"
)

const folder = "/seeds/"
const ext = ".sql"

func Seed(db database.Storer, filename string) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	src := fmt.Sprintf("%s%s%s%s", RootDir(), folder, filename, ext)
	dat, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(dat)); err != nil {
		if err = tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}
