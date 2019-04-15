package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFilter(t *testing.T) {
	wanted := []CharName{{0xAE, "REGISTERED SIGN"}}

	givenQuery := []string{"REGISTERED"}
	givenCharNames := []CharName{{0xAE, "REGISTERED SIGN"}, {0x23D, "LATIN CAPITAL LETTER L WITH BAR"}}

	got := Filter(givenCharNames, givenQuery)

	assert.Equal(t, wanted, got)
}

func TestFilter_query_cases(t *testing.T) {
	given := []CharName{
		{0x3c, "LESS-THAN SIGN"},
		{0xAE, "REGISTERED SIGN"},
		{0x23D, "LATIN CAPITAL LETTER L WITH BAR"},
	}

	testCases := []struct {
		description string
		query       []string
		want        []CharName
	}{
		{"Case Insensitive", []string{"registered"}, []CharName{{0xAE, "REGISTERED SIGN"}}},
		{"Whole Word Only", []string{"regis"}, []CharName{}},
		{"Not found", []string{"something that not exists"}, []CharName{}},
		{"Hyphenated Name", []string{"LESS"}, []CharName{{0x3c, "LESS-THAN SIGN"}}},
		{"Hyphenated Query", []string{"LESS-THAN"}, []CharName{{0x3c, "LESS-THAN SIGN"}}},
		{"Multiple Results", []string{"SIGN"}, []CharName{{0x3c, "LESS-THAN SIGN"}, {0xAE, "REGISTERED SIGN"}}},
		{"Multiple Queries Order Insensitive", []string{"SIGN", "LESS"}, []CharName{{0x3c, "LESS-THAN SIGN"}}},
		{"Empty Query", []string{}, []CharName{}},
		{"Duplicate Query Name", []string{"REGISTERED", "REGISTERED"}, []CharName{{0xAE, "REGISTERED SIGN"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			got := Filter(given, tc.query)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseUnicodeLine(t *testing.T) {
	want := CharName{0x39, "DIGIT NINE"}

	given := "0039;DIGIT NINE;Nd;0;EN;;9;9;9;N;;;;;"

	got := ParseUnicodeLine(given)

	assert.Equal(t, want, got)
}

func Example_Display() {
	given := []CharName{
		{0x3c, "LESS-THAN SIGN"},
		{0xAE, "REGISTERED SIGN"},
		{0x23D, "LATIN CAPITAL LETTER L WITH BAR"},
	}
	Display(given)
	// Output:
	// U+003C	<	LESS-THAN SIGN
	// U+00AE	®	REGISTERED SIGN
	// U+023D	Ƚ	LATIN CAPITAL LETTER L WITH BAR
}

func TestReadUnicodeData(t *testing.T) {
	want := []CharName{
		{'0', "DIGIT ZERO"},
		{'1', "DIGIT ONE"},
		{'2', "DIGIT TWO"},
	}

	got, err := ReadUnicodeData("UnicodeDataFixture.txt")

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestDownloadUnicodeFile(t *testing.T) {
	os.Remove("UnicodeData.txt")

	filename, err := DownloadUnicodeFile()

	got, err := ReadUnicodeData(filename)

	assert.FileExists(t, filename)
	assert.NoError(t, err)
	assert.Equal(t, CharName{0x20, "SPACE"}, got[32])
	assert.True(t, len(got) > 31000)
}
