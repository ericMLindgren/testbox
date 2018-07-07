package challenges

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Basic CRUD API

func GetById(id ID) (Challenge, error) {
	var c Challenge
	err := db.QueryRow("SELECT * FROM challenges WHERE id = ?", id).Scan(&c.ID, &c.Name, &c.ShortDesc, &c.LongDesc, &c.Tags, &c.SampleIO, &c.Cases)
	if err != nil {
		fmt.Printf("GetById error: %s", err.Error())
	}

	return c, err
}

func Insert(c *Challenge) (ID, error) {
	// fmt.Println("Insert not implemented")
	// should create a new record and return the id of said record
	res, err := db.Exec(
		"insert into challenges (name, short_desc, long_desc, tags, sampleIO, cases) values (?, ?, ?, ?, ?, ?);",
		c.Name, c.ShortDesc, c.LongDesc, c.Tags, c.SampleIO, c.Cases,
	)
	if err != nil {
		fmt.Printf("insert failed: %s", err.Error())
		return -1, err
	}

	id, _ := res.LastInsertId()
	// Set challenge objects ID to match TODO is this a good approach? seems side-effect-y
	c.ID = id
	fmt.Printf("success! entry key: %d\n", id)
	return id, err
}

func Update(id ID, c *Challenge) error {
	_, err := db.Exec(
		"UPDATE challenges SET name = ?, short_desc = ?, long_desc = ?, tags = ?, sampleIO = ?, cases = ? WHERE id = ?;",
		c.Name, c.ShortDesc, c.LongDesc, c.Tags, c.SampleIO, c.Cases, id)
	if err != nil {
		fmt.Printf("update error: %s\n", err.Error())
	}
	return err
}

func Delete(id ID) error {
	res, err := db.Exec("DELETE FROM challenges WHERE id = ?", id)
	if err != nil {
		fmt.Printf("delete error: %s\n", err.Error())
	}
	r, _ := res.RowsAffected()
	if r > 0 {
		fmt.Printf("entry %d deleted\n", id)
	} else {
		fmt.Printf("delete error: entry %d not found\n", id)
	}
	return nil
}

// Specialized retrieval methods

func GetRandom() Challenge {
	n := rand.Intn(countChallenges()) + 1
	id := int64(n)
	c, _ := GetById(id) // handle error TODO

	c, _ = GetById(1) // testing purposes only TODO WARNING
	return c
}

func GetByTag() {

}

func GetAll() map[int64]Challenge {
	// fmt.Println("Retrieving all entries...")
	rows, err := db.Query("SELECT * FROM challenges")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	results := make(map[int64]Challenge)
	for rows.Next() {
		var e Challenge
		err = rows.Scan(&e.ID, &e.Name, &e.ShortDesc, &e.LongDesc, &e.Tags, &e.SampleIO, &e.Cases)
		if err != nil {
			panic(err)
		}

		results[e.ID] = e
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}
	// fmt.Println("Entries retrieved")

	return results
}

// TODO handle graceful closing

var db *sql.DB

func OpenDB() {
	rand.Seed(time.Now().UTC().UnixNano())

	fmt.Print("Connecting to database...")
	var err error
	db, err = sql.Open("sqlite3", "./data/challenges.db")
	if err != nil {
		panic("failed to connect to database")
	}
	fmt.Printf("DB contains %d challenge(s)\n", countChallenges())

	err = resetChallengeTable()
	if err != nil {
		fmt.Printf("ERROR: Unable to reset challenge table: %s\n", err.Error())
		return
	}

	fmt.Println("inserting dummy challenge...")
	dummy1 := dummyChallenge()
	_, _ = Insert(dummy1)
}

func testDB() {
	fmt.Println("inserting dummy challenge...")
	dummy1 := dummyChallenge()
	id, _ := Insert(dummy1)
	_ = id

	c, _ := GetById(id)
	fmt.Printf("Retrieving challenge %d:\n%v\n", id, c)

	dummy2 := dummy1
	dummy2.Name = "Not so dumb"
	fmt.Println("Updating!")
	Update(id, dummy2)

	c, _ = GetById(id)
	fmt.Printf("Retrieving challenge %d:\n%v\n", id, c)

	fmt.Printf("challenge count: %d\n", countChallenges())
	fmt.Println("testing delete...")
	Delete(25)
	Delete(id)
	fmt.Printf("challenge count: %d\n", countChallenges())

	c, _ = GetById(id)
	fmt.Printf("Retrieving challenge %d:\n%v\n", id, c)
}

func CloseDB() {
	db.Close()
}

func countChallenges() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM challenges").Scan(&count)
	if err != nil {
		panic(err)
	}

	return count
}

func resetChallengeTable() error {
	err := removeChallengeTable()
	if err != nil {
		return err
	}

	err = initChallengesTable()
	if err != nil {
		return err
	}

	return nil
}

func removeChallengeTable() error {
	fmt.Println("Removing old challenges table...")
	if countChallenges() != 1 {
		return errors.New("Challenge table has non-trivial entries, delete file manually")
	}
	sqlStatement := `DROP TABLE challenges`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		fmt.Printf("error removing table: %s", err.Error())
	}
	fmt.Println("Table removed")
	return nil
}

func initChallengesTable() error {
	fmt.Println("Creating challenges table (if it doesn't exist)...")
	sqlStatement := `
		CREATE TABLE IF NOT EXISTS challenges (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		short_desc TEXT NOT NULL,
		long_desc TEXT,
		tags TEXT,
		sampleIO TEXT NOT NULL,
		cases TEXT NOT NULL
		);`

	_, err := db.Exec(sqlStatement)
	if err != nil {
		return err
	}

	fmt.Println("Challenge table created")
	return nil
}
