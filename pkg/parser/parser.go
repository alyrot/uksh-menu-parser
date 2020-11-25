package parser

//go:generate mockgen -source parser.go -destination ../../mocks/pkg/parser/parser.go

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	_ "image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/disintegration/imaging"
	"github.com/snabb/isoweek"
)

type Dish struct {
	Title       string
	Description string
	Price       string
	Kcal        string
	Type        string
	Date        time.Time
	colID       int
	rowID       int
}

type UKSHParserI interface {
	PDFToDishes(pdf []byte) ([]*Dish, error)
}

type UKSHParser struct{}

func (p *UKSHParser) PDFToDishes(pdf []byte) ([]*Dish, error) {
	return PDFToDishes(pdf)
}

func (d Dish) String() string {
	return fmt.Sprintf("Type: %v Title=%v Description=%v Price=%v Kcal=%v\n", d.Type, d.Title, d.Description, d.Price, d.Kcal)
}

/*
execInOutCMD, executes the program denoted by value with the provided flags and in as stdin.
If  returns the stdout of the program or an error. The stderr of the program is currently
not passed along
*/
func execInOutCMD(name string, flags []string, in []byte) ([]byte, error) {
	cmd := exec.Command(name, flags...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("attaching stdin failed: %v", err)
	}
	var stdInErr error
	go func() {
		defer stdin.Close()
		_, stdInErr = io.Copy(stdin, bytes.NewReader(in))
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("attaching stdout failed: %v", err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("attaching stderr failed: %v", err)
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start failed: %v", err)
	}

	buffer := bytes.NewBuffer(make([]byte, 0))
	if _, err := io.Copy(buffer, stdout); err != nil {
		return nil, fmt.Errorf("failed to read stdout: %v", err)
	}

	errMsg := bytes.NewBuffer(make([]byte, 0))
	if _, err := io.Copy(errMsg, stderr); err != nil {
		errMsg.WriteString("failed to get buffer")
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%v %s", err, errMsg)
	}
	if stdInErr != nil {
		return nil, fmt.Errorf("error copying to stdin: %v %s", err, errMsg)
	}
	return buffer.Bytes(), nil
}

/**
OCRImage, passes the image contained in img to tesseract with "-l deu" and returns
the text recognized by tesseract or an error
*/
func OCRImage(img []byte) (string, error) {
	out, err := execInOutCMD("tesseract", []string{"-l", "deu", "stdin", "stdout"}, img)
	if err != nil {
		return "", fmt.Errorf("OCRImage: %v", err)
	}
	return string(out), nil
}

/*
PDFToPng, uses "pdftoppm" to convert pdf to an img
*/
func PDFToPng(pdf []byte) ([]byte, error) {
	out, err := execInOutCMD("pdftoppm", []string{"-png"}, pdf)
	if err != nil {
		return nil, fmt.Errorf("PDFToPng: %v", err)
	}
	return out, nil
}

func PDFToText(pdf []byte) ([]byte, error) {
	outFile, err := ioutil.TempFile("", "*")
	if err != nil {
		return nil, fmt.Errorf("failed to create tmp output file: %v", err)
	}
	defer os.Remove(outFile.Name())
	defer outFile.Close()
	//outFilePath := filepath.Join(os.TempDir(),outFile.Name())
	_, err = execInOutCMD("pdftotext", []string{"-layout", "-", outFile.Name()}, pdf)
	if err != nil {
		return nil, fmt.Errorf("PDFToText: %v", err)
	}
	out, err := ioutil.ReadAll(outFile)
	if err != nil {
		return nil, fmt.Errorf("PDFTOText: failed to read output from %v: %v", outFile, err)
	}

	return out, nil
}

type imageContainer struct {
	img   *image.NRGBA
	colID int
	rowID int
}

/*
UKSHMenuToTiles, expects the bytes of png depicting the uksh menu plan and applies
whaky 2D arithmetic to cut in into tiles containing the single dishes
*/
func UKSHMenuToTiles(imgBytes []byte) ([]imageContainer, error) {
	img, err := imaging.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}
	const (
		//top left corner of the first tile
		xAnchor = 303
		yAnchor = 330

		//x dimension of tile
		tileWidth = 338
		//y dimension of tile
		tileHeight = 119
		//number of rows in table; row equals week day
		tileRowCount = 7
		//number of columns in each row, equals dishes
		tileColCount = 4 //4
	)

	var xTopLeft, yTopLeft, xBotRight, yBotRight int
	tiles := make([]imageContainer, 0, tileRowCount*tileColCount)
	for row := 0; row < tileRowCount; row++ {
		for col := 0; col < tileColCount; col++ {
			xTopLeft = xAnchor + (col * tileWidth)
			yTopLeft = yAnchor + (row * tileHeight)
			xBotRight = xTopLeft + tileWidth
			yBotRight = yTopLeft + tileHeight
			region := image.Rect(xTopLeft, yTopLeft, xBotRight, yBotRight)
			tile := imaging.Crop(img, region)
			tiles = append(tiles, imageContainer{
				img:   tile,
				colID: col,
				rowID: row,
			})
		}
	}

	return tiles, nil
}

func parseDish(text string) (*Dish, error) {
	d := &Dish{}
	lines := strings.Split(text, "\n")
	d.Title = lines[0]
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Trim(lines[i], "\n")
		lines[i] = strings.Trim(lines[i], " ")
		if strings.Contains(lines[i], "€") {
			startKcal := strings.IndexAny(lines[i], "k")
			if startKcal == -1 {
				d.Price = lines[i]
				d.Kcal = ""
			} else {
				d.Price = lines[i][:startKcal]
				d.Kcal = lines[i][startKcal:]
			}
			break //not interested in any dangling weird lines
		} else {
			d.Description += " "
			d.Description += lines[i]
		}
	}
	return d, nil
}

func TileToDish(tile []byte) (*Dish, error) {
	text, err := OCRImage(tile)
	if err != nil {
		return nil, fmt.Errorf("OCRImage: %v", err)
	}
	return parseDish(text)
}

/*
Table token in text version of pdf plan
*/
type token struct {
	offset int    //starting position in the respective line
	value  string //value
}

/*
Table header in the text version of pdf plan
*/
type column struct {
	offset int    //starting position in the respective line
	name   string //value
	id     int    //logical enumeration of the columns
}

/*
mathToColumn returns the column with column.offset closest to index
Assumes columns is sorted in ascending order
*/
func matchToColumn(index int, columns []*column) *column {
	for i := 0; i < len(columns)-1; i++ {
		if math.Abs(float64(index-columns[i].offset)) <= math.Abs(float64(index-columns[i+1].offset)) {
			return columns[i]
		}
	}

	return columns[len(columns)-1]
}

/*
splitAfterXWhitespaces, tokenizes l into sequences of words that are at most separated by maxWhiteSpaces
Example for maxWhitespaces = 2 "Pasta Rustikal mit Ofengemüse  Schnitzel extra groß" -> [Pasta Rustikal mit Ofengemüse,Schnitzel extra groß]
*/
func splitAfterXWhitespaces(l string, maxWhiteSpaces int) []*token {

	l = strings.Trim(l, "")

	colStart := -1
	whiteSpaces := 0
	res := make([]*token, 0)

	for i, c := range l {
		//look for start of new col
		if colStart == -1 {
			if unicode.IsLetter(c) {
				colStart = i
				whiteSpaces = 0
			}
		} else { //
			if unicode.IsSpace(c) {
				if whiteSpaces < maxWhiteSpaces {
					whiteSpaces++
				} else { //sufficient whitespaces since last char, end token
					s := strings.Trim(l[colStart:i], " \n")
					res = append(res, &token{
						offset: colStart,
						value:  s,
					})
					colStart = -1
				}
			} else { //reset white space counter on on whitespace char
				whiteSpaces = 0
			}
		}
	}
	//handle dangling token
	if colStart != -1 {
		res = append(res, &token{colStart, strings.Trim(l[colStart:], " \n")})
	}
	return res
}

/*
hasIgnorePrefix, is a helper that cuts some unwanted prefixes from a string
*/
func hasIgnorePrefix(s string) bool {
	ignorePrefixes := []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"}
	s = strings.Trim(s, " \n")
	for i := range ignorePrefixes {
		if strings.HasPrefix(s, ignorePrefixes[i]) {
			return true
		}
	}
	return false
}

/*
textToDish takes the output of pdftotext and parses it into dishes.
The iso week is interpreted for year
Cannot obtain price information
*/
func textToDishInYear(text []byte, year int) ([]*Dish, error) {

	lines := strings.Split(string(text), "\n")
	if count := len(lines); count <= 3 {
		return nil, fmt.Errorf("textToDish: input has not enough lines\n")
	}

	isoWeekRegexp := regexp.MustCompile("[0-9]{1,2}")
	anchorDate := time.Time{}
	for i := range lines {
		if strings.HasPrefix(strings.Trim(lines[i], " "), "Speiseplan Bistro") {
			weekNrBytes := isoWeekRegexp.Find([]byte(lines[i]))
			if weekNrBytes == nil {
				return nil, fmt.Errorf("textToDish: parsing error: failed to locate week number")
			}
			weekNr, err := strconv.ParseInt(string(weekNrBytes), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("textToDish: parsing error: failed to parse %s to week number: %v", weekNrBytes, err)
			}
			anchorDate = isoweek.StartTime(year, int(weekNr), time.Local)
		}
	}
	if anchorDate.Equal(time.Time{}) {
		return nil, fmt.Errorf("textToDish: parsing error: failed to locate week number")
	}

	//find Wochentag line or exit
	lineWochentag := -1
	for i := range lines {
		if strings.HasPrefix(strings.Trim(lines[i], " "), "Wochentag") {
			lineWochentag = i
		}
	}
	if lineWochentag == -1 {
		return nil, fmt.Errorf("textToDish: parsing error: \"Wochentag\" line not found")
	}

	//get whitespace offset of the headers in the lines
	headers := []string{"Wok Station", "Vegetarisch", "Gericht 2", "Gericht 3"}
	headerLine := lines[lineWochentag]
	columns := make([]*column, 0, len(headers))
	for i := range headers {
		index := strings.Index(headerLine, headers[i])
		if index == -1 {
			return nil, fmt.Errorf("textToDish: parsing error: failed to locate %v in \"Wochentag\" line", headers[i])
		}
		columns = append(columns, &column{offset: index, name: headers[i], id: i})
	}

	dishes := make([]*Dish, 0)

	//main parse loop look for line with kcal to seperate the "rows". Go backwards from there to parse dishes
	lastKcalLine := lineWochentag
	const kcal = "kcal"
	rowID := 0
	for i := lineWochentag + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], kcal) {
			startLine := lastKcalLine + 1
			kcalLine := i

			//parse textual description
			titles := splitAfterXWhitespaces(lines[startLine], 3)

			descriptions := make([][]*token, 0)
			for j := startLine + 1; j < kcalLine; j++ {
				if !hasIgnorePrefix(lines[j]) {
					descriptions = append(descriptions, splitAfterXWhitespaces(lines[j], 3))
				}
			}
			nutritionalValue := splitAfterXWhitespaces(lines[kcalLine], 3)

			//maps coldID to Dish if it exists
			tmp := make(map[int]*Dish)
			for j := range titles {
				id := matchToColumn(titles[j].offset, columns).id
				v, ok := tmp[id]
				if !ok {
					tmp[id] = &Dish{Title: titles[j].value}
				} else {
					v.Title = titles[j].value
				}
			}

			for j := range descriptions {
				for k := range descriptions[j] {
					id := matchToColumn(descriptions[j][k].offset, columns).id
					v, ok := tmp[id]
					if !ok {
						tmp[id] = &Dish{Description: descriptions[j][k].value}
					} else {
						if v.Description != "" {
							v.Description += ", "
						}
						v.Description += descriptions[j][k].value
						v.Description = strings.Trim(v.Description, " ,")
					}
				}
			}

			for j := range nutritionalValue {
				id := matchToColumn(nutritionalValue[j].offset, columns).id
				v, ok := tmp[id]
				if !ok {
					tmp[id] = &Dish{Kcal: nutritionalValue[j].value}
				} else {
					v.Kcal = nutritionalValue[j].value
				}
			}

			//filter entries with kcal to eliminate "Tageskarte" entries
			for colID, v := range tmp {
				if v.Kcal != "" {
					v.Type = columns[colID].name
					v.rowID = rowID
					v.colID = colID
					v.Date = anchorDate.AddDate(0, 0, rowID)
					dishes = append(dishes, v)
				}
			}

			//update for next round
			lastKcalLine = i
			rowID++
		}
	}

	return dishes, nil
}

/*
mergeTextAndOCR, contains the information with textToDish with the price information from
the OCR analysis
*/
func PDFToDishes(pdf []byte) ([]*Dish, error) {
	return PDFToDishesInYear(pdf, time.Now().In(time.Local).Year())
}

/*
mergeTextAndOCR, contains the information with textToDish with the price information from
the OCR analysis
*/
func PDFToDishesInYear(pdf []byte, year int) ([]*Dish, error) {

	text, err := PDFToText(pdf)
	if err != nil {
		return nil, fmt.Errorf("mergeTextAndOCR: %v", err)
	}

	pdfAsPNG, err := PDFToPng(pdf)
	if err != nil {
		return nil, fmt.Errorf("mergeTextAndOCR: %v", err)
	}

	tiles, err := UKSHMenuToTiles(pdfAsPNG)
	if err != nil {
		return nil, fmt.Errorf("mergeTextAndOCR: %v", err)
	}

	rowColPrice := make([][]string, 7)
	for i := range rowColPrice {
		rowColPrice[i] = make([]string, 4)
	}
	buf := new(bytes.Buffer)
	for i := range tiles {
		if err := png.Encode(buf, tiles[i].img); err != nil {
			return nil, fmt.Errorf("mergeTextAndOCR: conversion of tile (%v,%v) to []byte failed: %v", tiles[i].rowID, tiles[i].colID, err)
		}
		d, err := TileToDish(buf.Bytes())
		buf.Reset()
		if err != nil {
			return nil, fmt.Errorf("mergeTextAndOCR: tile (%v,%v): %v", tiles[i].rowID, tiles[i].colID, err)
		}
		rowColPrice[tiles[i].rowID][tiles[i].colID] = d.Price
	}

	dishes, err := textToDishInYear(text, year)
	if err != nil {
		return nil, fmt.Errorf("mergeTextAndOCR: %v", err)
	}

	for i := range dishes {
		dishes[i].Price = rowColPrice[dishes[i].rowID][dishes[i].colID]
	}

	return dishes, nil

}
