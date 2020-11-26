package main

import (
	"errors"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	menuCacheMock "github.com/alyrot/uksh-menu-parser/mocks/cmd/web"

	parserMock "github.com/alyrot/uksh-menu-parser/mocks/pkg/parser"

	"github.com/alyrot/uksh-menu-parser/pkg/parser"

	"github.com/golang/mock/gomock"
)

func TestExtractLinks(t *testing.T) {

	path := "../../testFiles/menuSite.html"
	realisticSiteBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		name          string
		in            []byte
		expectedLinks []string
		shouldFail    bool
	}

	//<a.+>Speiseplan Bistro.+<\/a>

	tests := []*testCase{
		{
			name: "Realistic Site",
			in:   realisticSiteBytes,
			expectedLinks: []string{
				"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+47.pdf",
				"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf",
			},
		},
		{
			name:          "No links",
			in:            []byte("you shall find no links here"),
			expectedLinks: nil,
			shouldFail:    false,
		},
		{
			name:          "Only one link",
			in:            []byte(`<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf" target="_blank">Speiseplan Bistro KW 48&nbsp;[pdf]</a>`),
			expectedLinks: []string{"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf"},
		},
		{
			name: "Unusual many links",
			in: []byte(
				`<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf" target="_blank">Speiseplan Bistro KW 48&nbsp;[pdf]</a>
				<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+49.pdf" target="_blank">Speiseplan Bistro KW 49&nbsp;[pdf]</a>
				<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+50.pdf" target="_blank">Speiseplan Bistro KW 50&nbsp;[pdf]</a>`),
			expectedLinks: []string{"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf",
				"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+49.pdf",
				"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+50.pdf"},
		},
	}

	//<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf" target="_blank">Speiseplan Bistro KW 48&nbsp;[pdf]</a>
	for _, v := range tests {
		func(tc *testCase) {
			t.Run(tc.name, func(t *testing.T) {
				got, err := extractLinks(tc.in)
				if err != nil {
					if !tc.shouldFail {
						t.Errorf("Unexpected error: %v\n", err)
					}
				} else {
					if tc.shouldFail {
						t.Errorf("Did not encounter expected error\n")
					}
					for i := range tc.expectedLinks {
						if tc.expectedLinks[i] != got[i] {
							t.Errorf("Expected %v got %v\n", tc.expectedLinks[i], got[i])
						}
					}
				}
			})
		}(v)
	}

}

/*
createDownloaderMock, creates a realistic mock for extractPDFsFromMenuSite that
returns [][]byte
*/
func createDownloaderMock(ctrl *gomock.Controller) (*menuCacheMock.MockDownloader, [][]byte, error) {
	//load dummy data for download mock
	dummySite, err := ioutil.ReadFile("../../testFiles/menuSite.html")
	if err != nil {
		return nil, nil, err
	}
	const pdf1URL = `https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+47.pdf`
	dummyPDF1, err := ioutil.ReadFile("../../testFiles/planKW47.pdf")
	if err != nil {
		return nil, nil, err
	}
	const pdf2URL = `https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf`
	dummyPDF2, err := ioutil.ReadFile("../../testFiles/planKW48.pdf")
	if err != nil {
		return nil, nil, err
	}

	downloadMock := menuCacheMock.NewMockDownloader(ctrl)
	downloadMock.EXPECT().Get(MenuBaseURL).Return(dummySite, nil)
	downloadMock.EXPECT().Get(pdf1URL).Return(dummyPDF1, nil)
	downloadMock.EXPECT().Get(pdf2URL).Return(dummyPDF2, nil)

	return downloadMock, [][]byte{dummyPDF1, dummyPDF2}, nil
}

func TestExtractPDFsFromMenuSite(t *testing.T) {
	ctrl := gomock.NewController(t)

	realisticIn, realisticExp, err := createDownloaderMock(ctrl)
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		name       string
		in         *menuCacheMock.MockDownloader
		exp        [][]byte
		shouldFail bool
	}

	threeDummyPDF := [][]byte{[]byte("dummy 1"), []byte("dummy 2"), []byte("dummy 3")}

	tests := []*testCase{
		{
			name:       "Realistic site",
			in:         realisticIn,
			exp:        realisticExp,
			shouldFail: false,
		},
		{
			name: "Unusual amount of links",
			in: func() *menuCacheMock.MockDownloader {
				d := menuCacheMock.NewMockDownloader(ctrl)
				d.EXPECT().Get(MenuBaseURL).Return([]byte(
					`<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf" target="_blank">Speiseplan Bistro KW 48&nbsp;[pdf]</a>
				<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+49.pdf" target="_blank">Speiseplan Bistro KW 49&nbsp;[pdf]</a>
				<a href="/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+50.pdf" target="_blank">Speiseplan Bistro KW 50&nbsp;[pdf]</a>`), nil)
				d.EXPECT().Get("https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf").Return(threeDummyPDF[0], nil)
				d.EXPECT().Get("https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+49.pdf").Return(threeDummyPDF[1], nil)
				d.EXPECT().Get("https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+50.pdf").Return(threeDummyPDF[2], nil)
				return d
			}(),
			exp:        threeDummyPDF,
			shouldFail: false,
		},
	}

	for _, v := range tests {
		func(tc *testCase) {
			t.Run(tc.name, func(t *testing.T) {

				got, err := extractPDFsFromMenuSite(tc.in)
				if err != nil {
					if !tc.shouldFail {
						t.Errorf("Unexpected Error: %v\n", err)
					}
				} else if tc.shouldFail {
					t.Errorf("Expected error got none\n")
				}

				if l := len(got); l != len(tc.exp) {
					t.Errorf("Expected %v PDFs got %v\n", len(tc.exp), l)
				}

				if !reflect.DeepEqual(got, tc.exp) {
					t.Errorf("Wrong PDFs returned\n")
				}
			})
		}(v)
	}
}

func TestMenuCache_GetMenu(t *testing.T) {
	ctrl := gomock.NewController(t)

	cachedDate := roundToDay(time.Now().In(time.Local))

	cachedDishes := []*parser.Dish{
		{Title: "Cached Dummy 1", Date: cachedDate},
		{Title: "Cached Dummy 2", Date: cachedDate},
	}

	uncachedDate := roundToDay(time.Now().In(time.Local).Add(24 * time.Hour))
	uncachedDishes := []*parser.Dish{
		{Title: "Uncached Dummy 1", Date: uncachedDate},
		{Title: "Uncached Dummy 2", Date: uncachedDate},
	}

	downloadMock, mockPDFs, err := createDownloaderMock(ctrl)
	if err != nil {
		t.Fatal(err)
	}

	parseMock := parserMock.NewMockUKSHParserI(ctrl)
	parseMock.EXPECT().PDFToDishes(mockPDFs[0]).Return(uncachedDishes, nil)
	parseMock.EXPECT().PDFToDishes(mockPDFs[1]).Return([]*parser.Dish{}, nil)

	mc := MenuCache{
		dateToDishes: map[time.Time][]*parser.Dish{cachedDate: cachedDishes},
		download:     downloadMock,
		parse:        parseMock,
		infoLog:      log.New(ioutil.Discard, "", 0),
		errorLog:     log.New(ioutil.Discard, "", 0),
	}

	res, err := mc.GetMenu(cachedDate)
	if err != nil {
		t.Errorf("Unexpected Error: %v", err)
	}
	if !reflect.DeepEqual(res, cachedDishes) {
		t.Errorf("Expected %v got %v\n", cachedDishes, res)
	}

	res, err = mc.GetMenu(uncachedDate)
	if err != nil {
		t.Errorf("Unexpected Error: %v", err)
	}

	if !reflect.DeepEqual(res, uncachedDishes) {
		t.Errorf("Expected %v got %v\n", uncachedDishes, res)
	}

	//test invalid date error
	for _, v := range []time.Time{cachedDate.Add(-24 * time.Hour), roundToDay(time.Now().Add(8 * 24 * time.Hour))} {
		_, err = mc.GetMenu(v)
		if err == nil {
			t.Errorf("Expected %v error but got none\n", invDateError)
		} else {
			if !errors.Is(err, invDateError) {
				t.Errorf("Expected %v error but got %v\n", invDateError, err)
			}
		}

	}

}
