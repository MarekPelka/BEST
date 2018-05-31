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
	"strconv"
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

func Pow(a, b int) int {
	p := 1
	for b > 0 {
		if b&1 != 0 {
			p *= a
		}
		b >>= 1
		a *= a
	}
	return p
}

func reduction(hash string, columnNumber int) string {
	// it generated random number 1-15 xD
	preety_random_hex := hash[0:1]
	preety_random_number, _ := strconv.ParseInt("0x"+preety_random_hex, 0, 8)
	passLength := int(preety_random_number)
	// Lets make better range, for example 6-10, number estimate wil be preety the same ?
	if passLength < 6 {
		passLength = passLength + 5
	} else if passLength > 10 {
		passLength = passLength - 5
	}
	// random numbers in 1-15 range
	random_numbers := make([]int, passLength)
	for i := 0; i < passLength; i++ {
		tmp_int64, _ := strconv.ParseInt("0x"+hash[i:i+1], 0, 8)
		random_numbers[i] = int(tmp_int64)
	}

	newPass := ""
	for i := 0; i < passLength; i++ {
		// Select from passwordcharacters index dependent on column
		// lets modulo columnNumber to not exceed int length and write something bad on memory (square of 2mld)
		randomChar := passwordCharacters[(Pow(random_numbers[i], 3)+Pow(columnNumber%200, 3))%len(passwordCharacters)]
		newPass += string(randomChar)
	}
	fmt.Printf("\nReduction %s -> %s\n", hash, newPass)
	return newPass
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

	tsql := fmt.Sprintf("UPDATE RainbowSchema.rainbow SET WORD=@Word WHERE HASH=@Hash; IF @@ROWCOUNT = 0 INSERT INTO RainbowSchema.rainbow (HASH, WORD) VALUES (@Hash, @Word);")

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

func iter_to_the_last_hash(start_column int, start_hash string) string {
	hash := start_hash
	//fmt.Printf("Iterating over %d -> %d\n", start_column, width-1)
	for iter_column := start_column; iter_column < width-1; iter_column++ {
		reducted_iter := reduction(hash, iter_column)
		hash = hashString(reducted_iter)
		//fmt.Printf("%d:Hashing %s -> %s\n", iter_column, reducted_iter, hash)
		//fmt.Printf("hash %s, ", hash)
	}

	return hash
}

func find(hash string) string {

	numberOfIter := 0
	startWord := ""
	r := selectFromTable(hash)

	if r != "" {
		startWord = r
	} else {
		first_hash := hash
		for bet_on_column := width - 2; bet_on_column >= 0; bet_on_column-- {
			hash = first_hash
			fmt.Printf("\n\n\nBetting on %d\n", bet_on_column)
			//fmt.Printf("Hashing:%s\n", hash)
			hash = iter_to_the_last_hash(bet_on_column, hash)
			r = selectFromTable(hash)

			if r != "" {
				numberOfIter = bet_on_column
				fmt.Printf("\nFoud in db %s -> %s\n", hash, r)
				startWord = r
				break
			}
		}
	}

	for i := 0; i < numberOfIter; i++ {

		fmt.Printf("\n%d Starting word: %s", i, startWord)
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
