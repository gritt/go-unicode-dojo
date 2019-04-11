package main

import (
	"fmt"
	"github.com/standupdev/strset"
	"strconv"
	"strings"
)

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

	return CharName{rune(code), fields[1] }
}


func Display(names []CharName) {
	for _, charName := range names {
		fmt.Printf("%U\t%c\t%s\n", charName.char, charName.char, charName.name)
	}
}