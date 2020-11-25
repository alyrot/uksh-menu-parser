package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/disintegration/imaging"
)

func TestOCRImage(t *testing.T) {

	t.Parallel()

	path := "../../testFiles/ocrTestImg.png"
	imgBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read  input from %v: %v\n", path, err)
	}

	text, err := OCRImage(imgBytes)

	if err != nil {
		t.Errorf("Unexpected error %v\n", err)
	}

	if !strings.Contains(text, "line 1") || !strings.Contains(text, "line 2") {
		t.Errorf("Expect OCR to contain \"line 1\" and \"line 2\" but got %v", text)
	}

	if text == "line 1\nline2" {
		t.Errorf("formatting not as expect\n")
	}

}

func TestPDFToPng(t *testing.T) {
	t.Parallel()
	path := "../../testFiles/planKW47.pdf"
	pdfBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read input %v: %v\n", path, err)
	}

	img, err := PDFToPng(pdfBytes)

	if err != nil {
		t.Errorf("Unexpected error %v\n", err)
	}

	if len(img) == 0 {
		t.Errorf("Resulting image is empty")
	}
	if err := ioutil.WriteFile("../../testFiles/planKW47.png", img, os.ModePerm); err != nil {
		t.Errorf("Failed to write output for visual inspection: %v\n", err)
	}
}

func TestPDFToText(t *testing.T) {
	t.Parallel()
	path := "../../testFiles/planKW47.pdf"
	pdfBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read input %v: %v", path, err)
	}

	text, err := PDFToText(pdfBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(text) == 0 {
		t.Errorf("unexpected output length 0\n")
	}
	//not sure what else to check. we could prepare a dummy pdf whose content is easily checkable

}

func TestUkshMenuToTiles(t *testing.T) {
	t.Parallel()
	//inPath := "../../testFiles/planKW47.png"

	inPath := "../../testFiles/prepared-planKW48.png"
	imgBytes, err := ioutil.ReadFile(inPath)
	if err != nil {
		t.Fatalf("Failed to read input from %v: %v", inPath, err)
	}

	tiles, err := UKSHMenuToTiles(imgBytes)
	if err != nil {
		t.Errorf("Got unexpected error: %v\n", err)
		t.FailNow()
	}

	if len(tiles) != 28 {
		t.Errorf("Unepexted number of results, wanted 28 got %v\n", len(tiles))
	}

	baseOutPath := "../../testFiles/testOut"
	//write out for manual inspection
	if _, err := os.Stat(baseOutPath); os.IsNotExist(err) {
		if err := os.Mkdir(baseOutPath, os.ModePerm); err != nil {
			t.Fatalf("Failed to create dir %v : %v", baseOutPath, err)

		}
	}
	for i := range tiles {
		outPath := fmt.Sprintf("%v/tile-%02d", baseOutPath, i)
		outFile, err := os.Create(outPath)
		if err != nil {
			t.Fatalf("Failed to create output file %v: %v", outPath, err)
		}
		defer outFile.Close()

		if err := imaging.Encode(outFile, tiles[i].img, imaging.PNG); err != nil {
			t.Errorf("Unexepected error encoding tile %v: %v", i, err)
			t.FailNow()
		}
	}
}

func TestTileToDish(t *testing.T) {
	t.Parallel()
	tileBasePath := "../../testFiles/preparedTiles"
	for i := 0; i < 28; i++ {
		tilePath := fmt.Sprintf("%v/tile-%02d", tileBasePath, i)
		imgBytes, err := ioutil.ReadFile(tilePath)
		if err != nil {
			t.Fatalf("Failed to read infile %v: %v\n", tilePath, err)
		}

		_, err = TileToDish(imgBytes)
		if err != nil {
			t.Errorf("Unexpected Error")
		}
		//t.Errorf("dish %v is %s\n",i,dish)

	}

}

/*
dishEqual, is a helper for TestTextToDish that checks if two dishes are equal
*/
func dishEqual(a, b *Dish, includingPrice bool) bool {
	res := a.Type == b.Type && a.Description == b.Description && a.Kcal == b.Kcal && a.rowID == b.rowID && a.colID == b.colID
	res = res && a.Date.Year() == b.Date.Year() && a.Date.Month() == b.Date.Month() && a.Date.Day() == b.Date.Day()
	return res && (!includingPrice || (a.Price == b.Price))

}

type dishTest struct {
	d     *Dish
	found bool
}

/*
Common test data for PDFToDishes and textToDish
*/
var sampleDishes = []*dishTest{
	{
		d: &Dish{
			Title:       "Pasta-Pfanne",
			Description: "mit Hähnchenfleisch",
			Price:       "€ 4,80 / € 6,00",
			Kcal:        "kcal 528 / kJ 2212",
			Type:        "Wok Station",
			Date:        time.Date(2020, 11, 16, 0, 0, 0, 0, time.Local),
			colID:       0,
			rowID:       0,
		},
	},
	{
		d: &Dish{
			Title:       "Rumpsteak",
			Description: "Champignon-Zwiebelgemüse, Bratkartoffeln und Kräuterbutter",
			Price:       "€ 5,90 / € 7,38",
			Kcal:        "kcal 879 / kJ 3683",
			Type:        "Gericht 2",
			Date:        time.Date(2020, 11, 22, 0, 0, 0, 0, time.Local),
			colID:       2,
			rowID:       6,
		},
	},
	{
		d: &Dish{
			Title:       "gebratenes Kabeljaufilet",
			Description: "mit Rahmwirsing, und Petersilienkartoffeln",
			Price:       "€ 4,90 / € 6,13",
			Kcal:        "kcal 429 / kJ 1797",
			Type:        "Gericht 3",
			Date:        time.Date(2020, 11, 18, 0, 0, 0, 0, time.Local),
			colID:       3,
			rowID:       2,
		},
	},
}

func TestPDFToDishes(t *testing.T) {
	t.Parallel()

	path := "../../testFiles/planKW47.pdf"
	pdfBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to open test input %v: %v\n", path, err)
	}

	dishes, err := PDFToDishesInYear(pdfBytes, 2020)
	if err != nil {
		t.Errorf("Unexpected error: %v\n", err)
		t.FailNow()
	}
	for i := range dishes {
		for j := range sampleDishes {
			if dishEqual(sampleDishes[j].d, dishes[i], true) {
				sampleDishes[j].found = true
			}
		}
	}

	for j := range sampleDishes {
		if !sampleDishes[j].found {
			t.Errorf("Did not find dish %v in output\n", *sampleDishes[j].d)
		}
	}

}

func TestTextToDish(t *testing.T) {
	t.Parallel()

	path := "../../testFiles/planKW47.txt"
	text, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to open test input %v: %v\n", path, err)
	}

	dishes, err := textToDishInYear(text, 2020)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	}

	if len(dishes) == 0 {
		t.Fatalf("Unexpected empty result\n")
	}

	for i := range dishes {
		for j := range sampleDishes {
			if dishEqual(sampleDishes[j].d, dishes[i], false) {
				sampleDishes[j].found = true
			}
		}
	}

	for j := range sampleDishes {
		if !sampleDishes[j].found {
			t.Errorf("Did not find dish %v in output\n", *sampleDishes[j].d)
		}
	}

}

func TestColumnifyLine(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name string
		in   string
		exp  []token
	}

	in := []testCase{
		{
			name: "Test 1",
			in:   `                         Pasta-Pfanne                                 Ofenkartoffel                   Bauernhacksteak Cordon Bleu                   Steckrüben-Kartoffeleintopf`,
			exp: []token{{
				offset: 25,
				value:  "Pasta-Pfanne",
			},
				{
					offset: 70,
					value:  "Ofenkartoffel",
				},
				{
					offset: 102,
					value:  "Bauernhacksteak Cordon Bleu",
				},
				{
					offset: 148,
					value:  "Steckrüben-Kartoffeleintopf",
				},
			},
		},
		/*{
			value: "Test 2",
			in: `                       mit Hähnchenfleisch                 mit Sour Creme, Spinat-Kürbisgemüse             Schwarzwurzelgmüse                   mit Kartoffelwürfel, Kohlwurstscheiben`,
			exp: []string{"mit Hähnchenfleisch","mit Sour Creme", "Spinat-Kürbisgemüse","Schwarzwurzelgmüse","mit Kartoffelwürfel, Kohlwurstscheiben"},
		},
		*/
	}

	for i := range in {
		func(i int) {
			t.Run(in[i].name, func(t *testing.T) {
				res := splitAfterXWhitespaces(in[i].in, 3)
				if len(res) != len(in[i].exp) {
					t.Errorf("Expected %v\ngot %v\n", in[i].exp, res)
				} else {
					for j := range res {
						if res[j].offset != in[i].exp[j].offset || res[j].value != in[i].exp[j].value {
							t.Errorf("Missmatch in element %v: expected (%v,%v) got (%v,%v)\n", i, in[i].exp[j].offset, in[i].exp[j].value, res[i].offset, res[i].value)
						}
					}
				}
			})
		}(i)
	}

}
