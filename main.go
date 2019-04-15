package main

import (
	"bufio"
	"fmt"
	"github.com/standupdev/strset"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const UnicodeDataURI string = "http://www.unicode.org/Public/UNIDATA/UnicodeData.txt"
const UnicodeDataFilename string = "UnicodeData.txt"

type CharName struct {
	char rune
	name string
}

func Filter(c []CharName, query []string) []CharName {
	results := []CharName{}

	if len(query) < 1 {
		return results
	}

	queryTerms := strset.MakeFromText(strings.ReplaceAll(strings.ToUpper(strings.Join(query, " ")), "-", " "))

	for _, charName := range c {
		nameSet := strset.MakeFromText(strings.ReplaceAll(charName.name, "-", " "))
		if queryTerms.SubsetOf(nameSet) {
			results = append(results, charName)
		}
	}

	return results
}

func ParseUnicodeLine(unicodeLine string) CharName {
	fields := strings.Split(unicodeLine, ";")

	code, _ := strconv.ParseInt(fields[0], 16, 32)

	return CharName{rune(code), fields[1]}
}

func Display(names []CharName) {
	for _, charName := range names {
		fmt.Printf("%U\t%c\t%s\n", charName.char, charName.char, charName.name)
	}
}

func ReadUnicodeData(filename string) ([]CharName, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	charNames := []CharName{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		charNames = append(charNames, ParseUnicodeLine(line))
	}

	return charNames, nil
}

func DownloadUnicodeFile() (string, error) {
	resp, err := http.Get(UnicodeDataURI)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.OpenFile(UnicodeDataFilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)

	return UnicodeDataFilename, nil
}
