package main

import (
	_ "github.com/denisenkom/go-mssqldb"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
	"flag"
	"bufio"
)

var server = "0.0.0.0"
var port = 1401
var user = "sa"
var password = "Microsoft2017"
var database = "HashDB"

var db *sql.DB

var wg sync.WaitGroup

var width = 1000

var rows = make(chan row)

var generate bool
var findPassword bool
var passwordFilename = "darkweb2017-top10000.txt"
var usage = "<FILENAME> - Generate rainbowtable with <FILENAME> as starting vectors for rows. Default FILENAME = " + passwordFilename

type row struct {
	word string
	hash string
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func reduction(h string) string {
	return h
}

//Have to end with empty line
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func appendToTable(r row) {

	defer wg.Done()
	ctx := context.Background()
	if db == nil {
		log.Fatal("What?")
	}
	err := db.PingContext(ctx)
	if err != nil {
		log.Fatal("Error pinging database: " + err.Error())
	}

	tsql := fmt.Sprintf("INSERT INTO RainbowSchema.rainbow (HASH, WORD) VALUES (@Hash, @Word);")

	_, err = db.ExecContext(
		ctx,
		tsql,
		sql.Named("Hash", r.hash),
		sql.Named("Word", r.word))

	if err != nil {
		log.Fatal("Error inserting new row: " + err.Error())
	}
}

func selectFromTable(hash string) string {

	ctx := context.Background()
	if db == nil {
		log.Fatal("What?")
	}
	err := db.PingContext(ctx)
	if err != nil {
		log.Fatal("Error pinging database: " + err.Error())
	}

	rows, err := db.Query("SELECT WORD FROM RainbowSchema.rainbow WHERE HASH='" + hash + "'")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var word string
	for rows.Next() {
		if err := rows.Scan(&word); err != nil {
			log.Fatal(err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return word
}

func find(hash string) string {

	var startWord string
	numberOfIter := 0
	for numberOfIter < width {
		r := selectFromTable(hash)

		if r != "" {
			startWord = r
			break
		}
		hash = reduction(hash)
		hash = hashString(hash)

		numberOfIter++
	}

	for i := 0; i < width - numberOfIter - 1; i++ {
		startWord = hashString(startWord)
		startWord = reduction(startWord)
	}
	return startWord
}

func init() {

	flag.BoolVar(&generate, "generate", false, usage)
	flag.BoolVar(&generate, "g", false, usage + " (shorthand)")
	flag.BoolVar(&findPassword, "f", false, "<HASH> - Find password for given <HASH>")
}

func openDBConnection() {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)
	var err error

	// Create connection pool
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	fmt.Printf("Connected!\n")
}

func openFile() *os.File {
	file, err := os.Open(passwordFilename)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func generatePass() {
	if len(os.Args) == 3 {
		passwordFilename = os.Args[2]
	}
	fmt.Printf("Using: %s\n", passwordFilename)
	openDBConnection()
	defer db.Close()
	file := openFile()
	defer file.Close()

	numberOfLines, _ := lineCounter(file)
	fmt.Printf("Numer of lines: %d\n", numberOfLines)
	wg.Add(numberOfLines)

	file.Seek(0, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		go func(w string) {
			var h string
			var subWord = w
			for i := 0; i < width; i++ {
				h = hashString(subWord)
				subWord = reduction(h)
			}

			rows <- row{w, h}
		}(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	go func() {
		for i := range rows {
			appendToTable(i)
		}
	}()

	wg.Wait()
}

func main() {

	flag.Parse()

	t1 := time.Now()
	fmt.Printf("Start time: %s\n", t1)

	switch {
	case generate:
		generatePass()
	case findPassword && len(os.Args) == 3:
		openDBConnection()
		fmt.Printf("Found password: %s\n", find(os.Args[2])) //"f82a7d02e8f0a728b7c3e958c278745cb224d3d7b2e3b84c0ecafc5511fdbdb7" --> sould return "password!"
	default:
		flag.Usage()
		os.Exit(0)
	}

	t2 := time.Since(t1)
	fmt.Printf("The query took: %s\n", t2)
}
