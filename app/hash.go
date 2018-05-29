package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	server   = "0.0.0.0"
	port     = 1401
	user     = "sa"
	password = "Microsoft2017"
	database = "HashDB"

	db *sql.DB

	wg sync.WaitGroup

	width = 100

	rows = make(chan row)

	generate         bool
	findPassword     bool
	passwordFilename = "darkweb2017-top10000.txt"
	usage            = "<FILENAME> - Generate rainbowtable with <FILENAME> as starting vectors for rows. Default FILENAME = " + passwordFilename

	defaultPassMinLength = 6
	defaultPassMaxLength = 12
	defaultLower         = "abcdefghijklmnopqrstuvwxyz"
	defaultUpper         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	defaultNumbers       = "0123456789"
	passwordCharacters   = defaultUpper + defaultLower + defaultNumbers
)

type row struct {
	word string
	hash string
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func reduction(h string, columnNumber int) string {
	inputBytes := []byte(h)
	newPass := ""
	for i := 0; i < random(defaultPassMinLength, defaultPassMaxLength); i++ {
		randomIndex := inputBytes[(i+columnNumber)%len(inputBytes)]
		generatedChar := passwordCharacters[int(randomIndex)%len(passwordCharacters)]
		newPass += string(generatedChar)
		//newPass.WriteByte(generatedChar)
	}
	fmt.Printf("%d: %s\n", columnNumber, newPass)

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

	r := selectFromTable(hash)
	if r != "" {
		startWord = r
	} else {
		for bet_on_column := width - 1; bet_on_column > 0; bet_on_column -- {
			for iter_column := bet_on_column; iter_column < width; iter_column ++ {
				hash = reduction(hash, iter_column)
				hash = hashString(hash)
			}
			numberOfIter++
			r = selectFromTable(hash)
			if r != "" {
				startWord = r
				break
			}
		}
	}

	for i := 0; i < width - numberOfIter ; i++ {
		startWord = hashString(startWord)
		startWord = reduction(startWord, i)
	}
	return startWord
}

func init() {

	flag.BoolVar(&generate, "generate", false, usage)
	flag.BoolVar(&generate, "g", false, usage+" (shorthand)")
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
	fmt.Printf("Numer of lines in provided pass dict file: %d\n", numberOfLines)
	wg.Add(numberOfLines)

	file.Seek(0, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Run for single line
		go func(w string) {
			var h string
			var subWord = w
			for i := 0; i < width; i++ {
				h = hashString(subWord)
				subWord = reduction(h, i)
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
		os.Exit(1)
	}

	t2 := time.Since(t1)
	fmt.Printf("The query took: %s\n", t2)
}
