package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	width                = 8
	h                    = ""
	preety_random_char   = ""
	preety_random_number = ""
	defaultPassMinLength = 6
	defaultPassMaxLength = 12
	defaultLower         = "abcdefghijklmnopqrstuvwxyz"
	defaultUpper         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	defaultNumbers       = "0123456789"
	passwordCharacters   = defaultUpper + defaultLower + defaultNumbers
)

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
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

func reduction(h string, columnNumber int) string {
	// it generated random number 1-15 xD
	preety_random_hex := h[0:1]
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
		tmp_int64, _ := strconv.ParseInt("0x"+h[i:i+1], 0, 8)
		random_numbers[i] = int(tmp_int64)
	}

	newPass := ""
	for i := 0; i < passLength; i++ {
		// Select from passwordcharacters index dependent on column
		// lets modulo columnNumber to not exceed int length and write something bad on memory (square of 2mld)
		randomChar := passwordCharacters[(Pow(random_numbers[i], 3)+Pow(columnNumber%200, 3))%len(passwordCharacters)]
		newPass += string(randomChar)
	}
	fmt.Printf("\nReduction %s -> %s\n", h, newPass)
	return newPass
}

func selectFromTable(h string) string {
	//fmt.Printf("Selecting  %s\n", h)
	if h == "6c8d890e11462dec081b5f382ff8c2eac7a16aeccee0190315d27813f4e00dee" {
		return "password"
	} else {
		return ""
	}
}

func iter_to_the_last_hash(start_column int, start_hash string) string {
	hash := start_hash
	fmt.Printf("Iterating over %d -> %d\n", start_column, width-1)
	for iter_column := start_column; iter_column < width-1; iter_column++ {
		reducted_iter := reduction(hash, iter_column)
		hash = hashString(reducted_iter)
		fmt.Printf("%d:Hashing %s -> %s\n", iter_column, reducted_iter, hash)
		//fmt.Printf("hash %s, ", hash)
	}

	return hash
}

func find(hash string) string {
	numberOfIter := 0
	startWord := ""
	r := ""

	if r != "" {
		startWord = r
	} else {
		first_hash := hash
		for bet_on_column := width - 2; bet_on_column >= 0; bet_on_column-- {
			hash = first_hash
			fmt.Printf("\n\n\nBetting on %d\n", bet_on_column)
			//fmt.Printf("\nFirst hash %s;\n", hash)
			//fmt.Printf("Hashing:%s\n", hash)
			hash = iter_to_the_last_hash(bet_on_column, hash)
			numberOfIter++
			//fmt.Printf("\n Last iter hash: %s, ", hash)
			r = selectFromTable(hash)

			if r != "" {
				println("KURWA")
				startWord = r
				break
			} else {
				fmt.Printf("Not in db:%s\n", hash)
			}
		}
	}

	for i := 0; i < width-numberOfIter-1; i++ {
		fmt.Printf("\nStarting word: %s", startWord)
		startWord = hashString(startWord)
		startWord = reduction(startWord, i)
	}
	return startWord
}

func main() {
	password := "password"
	for i := 0; i < width; i++ {
		h = hashString(password)
		println(i, password, h)
		password = reduction(h, i)
	}

	search := "40a5cd87cca586798f6fca96d76bec85723a08141def2ede14bd8d72ee016db9"
	password = find(search)
	println()
	println(password)
}
