package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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
		{0x3C, "LESS-THAN SIGN"},
		{0xAE, "REGISTERED SIGN"},
		{0x23D, "LATIN CAPITAL LETTER L WITH BAR"},
	}

	testCases := []struct {
		description string
		query       []string
		want        []CharName
	}{
		{"Should match case insensitive", []string{"registered"}, []CharName{{0xAE, "REGISTERED SIGN"}}},
		{"Should match whole words only", []string{"regis"}, []CharName{}},
		{"Should not found something that not exists", []string{"something that not exists"}, []CharName{}},
		{"Should match with hyphenated words", []string{"LESS"}, []CharName{{0x3C, "LESS-THAN SIGN"}}},
		{"Should match with hyphenated query", []string{"LESS-THAN"}, []CharName{{0x3C, "LESS-THAN SIGN"}}},
		{"Should return multiple results", []string{"SIGN"}, []CharName{{0x3C, "LESS-THAN SIGN"}, {0xAE, "REGISTERED SIGN"}}},
		{"Should be multiple queries order insensitive", []string{"SIGN", "LESS"}, []CharName{{0x3C, "LESS-THAN SIGN"}}},
		{"Should return empty for empty query", []string{}, []CharName{}},
		{"Should match one when query is duplicated", []string{"REGISTERED", "REGISTERED"}, []CharName{{0xAE, "REGISTERED SIGN"}}},
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
		{0x3C, "LESS-THAN SIGN"},
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

func TestSearchUnicodeDataHandler_query_cases(t *testing.T) {
	testCases := []struct {
		description  string
		query        []string
		wantResponse string
		wantStatus   int
	}{
		{
			"Should return error when null query is given",
			nil,
			`{"status":"error","message":"Empty query given","charNames":null}`,
			400,
		},
		{
			"Should return single result when given query matches one char name",
			[]string{"SMALL LETTER TURNED DELTA"},
			`{"status":"success","message":"Found these results for your search","charNames":[{"char":397,"name":"LATIN SMALL LETTER TURNED DELTA"}]}`,
			200,
		},
		{
			"Should return multiple results when given query matches multiple char names",
			[]string{"DESKTOP"},
			`{"status":"success","message":"Found these results for your search","charNames":[{"char":128421,"name":"DESKTOP COMPUTER"},{"char":128468,"name":"DESKTOP WINDOW"}]}`,
			200,
		},
		{
			"Should return single result when given multiple matching queries and queries should refine the search",
			[]string{"DESKTOP", "COMPUTER"},
			`{"status":"success","message":"Found these results for your search","charNames":[{"char":128421,"name":"DESKTOP COMPUTER"}]}`,
			200,
		},
		{
			"Should return empty result when given query dont match any char",
			[]string{"OCTOPUS CAT"},
			`{"status":"success","message":"Could not find any results for the given query","charNames":[]}`,
			200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			given := buildGETRequestWithQuery(t, tc.query)

			responseRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(SearchUnicodeDataHandler)
			handler.ServeHTTP(responseRecorder, given)

			assert.Equal(t, tc.wantResponse, responseRecorder.Body.String(), tc.description)
			assert.Equal(t, tc.wantStatus, responseRecorder.Code)
		})
	}
}

func buildGETRequestWithQuery(t *testing.T, query []string) *http.Request {

	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	if query == nil {
		return request
	}

	q := request.URL.Query()
	for _, param := range query {
		q.Add("query", param)
	}

	request.URL.RawQuery = q.Encode()

	return request
}
