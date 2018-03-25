// updateSchema.go: utility for updating the schema of jerver,
// reads the current version of the `jerver` schema from the
// `schemaVersion` table and applies any patch files (named
// <<version-number>>.sql) in the current directory to bring
// the schema up to date. If the jerver database does not exist,
// will create it and run patches from beginning. Should be run
// by a user with create privileges on postgres
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

var db *sql.DB

func init() {
	connStr := "user=jerver dbname=jerver sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func getHighestPatchNum() (int, error) {
	files, err := filepath.Glob("*.sql")

	if err != nil {
		return -1, err
	}

	highestVersion := -1
	for _, file := range files {
		patchNumber, err := strconv.Atoi(file[:len(file)-4])
		if err != nil {
			continue
		}
		if patchNumber > highestVersion {
			highestVersion = patchNumber
		}
	}

	if highestVersion == -1 {
		return -1, errors.New("No valid named SQL files found! (Should be named 0.sql, 1.sql etc)")
	}

	return highestVersion, nil
}

func createDB() {
	cmd := exec.Command("createdb", "jerver")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func getCurrentVersion() (int, error) {
	var ret int
	err := db.QueryRow(`
    SELECT
        schemaVersion 
    FROM
        schemaVersion
    `).Scan(&ret)

	if err != nil {
		switch err.(*pq.Error).Code {
		case "42P01": // table schema version does not exist
			return -1, nil
		case "3D000": // database does not exist
			createDB()
			return -1, nil
		case "28000": // jerver user does not exist
			createDB()
			return -1, nil // no need to unduly worry, the jerver user is created in 0.sql
		default:
			return -1, err
		}
	}

	return ret, nil
}

func writeSchemaVersion(version int) error {
	_, err := db.Exec("UPDATE schemaVersion SET schemaVersion = $1", version)
	return err
}

func main() {
	highestVersion, err := getHighestPatchNum()

	if err != nil {
		log.Fatal(err)
	}

	currentVersion, err := getCurrentVersion()

	if err != nil {
		log.Fatal(err)
	}

	for i := currentVersion + 1; i <= highestVersion; i++ {
		cmd := exec.Command("psql", "jerver", "-f", fmt.Sprintf("%d.sql", i))
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}

		err = cmd.Start()

		if err != nil {
			writeSchemaVersion(i - 1)
			log.Fatal(fmt.Sprintf("Error executing script %d.sql", i))
		}

		errorOp, _ := ioutil.ReadAll(stderr)
		matched, err := regexp.MatchString("psql.*ERROR", string(errorOp))
		if matched {
			writeSchemaVersion(i - 1)
			log.Fatal(fmt.Sprintf("Error executing script %d.sql: %s", i, errorOp))
		}

		matched, err = regexp.MatchString("psql.*FATAL", string(errorOp))
		if matched {
			writeSchemaVersion(i - 1)
			log.Fatal(fmt.Sprintf("Error executing script %d.sql: %s", i, errorOp))
		}

		if err := cmd.Wait(); err != nil {
			writeSchemaVersion(i - 1)
			log.Fatal(fmt.Sprintf("Error executing script %d.sql: %s", i, errorOp))
		}
	}

	err = writeSchemaVersion(highestVersion)
	if err != nil {
		log.Fatal(err)
	}
}
