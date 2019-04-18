package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
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
	Char rune   `json:"char"`
	Name string `json:"name"`
}

type CharNameResponse struct {
	Status    string     `json:"status"`
	Message   string     `json:"message"`
	CharNames []CharName `json:"charNames"`
}

func Filter(c []CharName, query []string) []CharName {
	results := []CharName{}

	if len(query) < 1 {
		return results
	}

	queryTerms := strset.MakeFromText(strings.ReplaceAll(strings.ToUpper(strings.Join(query, " ")), "-", " "))

	for _, charName := range c {
		nameSet := strset.MakeFromText(strings.ReplaceAll(charName.Name, "-", " "))
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
		fmt.Printf("%U\t%c\t%s\n", charName.Char, charName.Char, charName.Name)
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

func SearchUnicodeDataHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		newErrorResponse(w, errors.New("Empty query given"), http.StatusBadRequest)
		return
	}

	searchTerms := query["query"]

	unicodeData, err := ReadUnicodeData(UnicodeDataFilename)
	if err != nil {
		DownloadUnicodeFile()
		unicodeData, err = ReadUnicodeData(UnicodeDataFilename)
		if err != nil {
			newErrorResponse(w, errors.New("Failed to read unicode data"), http.StatusInternalServerError)
			return
		}
	}

	charNames := Filter(unicodeData, searchTerms)

	resp := CharNameResponse{}
	resp.Status = "success"

	resp.Message = "Found these results for your search"
	if len(charNames) == 0 {
		resp.Message = "Could not find any results for the given query"
	}

	resp.CharNames = charNames
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		newErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func newErrorResponse(w http.ResponseWriter, err error, code int) {
	resp := CharNameResponse{}
	resp.Status = "error"
	resp.Message = err.Error()
	jsonResp, err := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonResp)
}
