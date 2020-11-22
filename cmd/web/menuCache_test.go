package main

import (
	"errors"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	parserMock "github.com/alyrot/menuToText/mocks/pkg/parser"

	"github.com/alyrot/menuToText/pkg/parser"

	menuCacheMock "github.com/alyrot/menuToText/mocks/cmd/web"

	"github.com/golang/mock/gomock"
)

func TestExtractLinks(t *testing.T) {

	path := "../../testFiles/menuSite.html"
	siteBytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expectedLinks := []string{
		"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+47.pdf",
		"https://www.uksh.de/uksh_media/Speisepl%C3%A4ne/L%C3%BCbeck+_+UKSH_Bistro/Speiseplan+Bistro+KW+48.pdf",
	}

	links, err := extractLinks(siteBytes)
	if err != nil {
		t.Errorf("Unexpected error: %v\n", err)
	} else {
		for i := range expectedLinks {
			if expectedLinks[i] != links[i] {
				t.Errorf("Expected %v got %v\n", expectedLinks[i], links[i])
			}
		}
	}
}

/*
createDownloaderMock, creates a realistic mock for extractPDFsFromMenuSite that
returns [][]byte
*/
func createDownloaderMock(ctrl *gomock.Controller) (*menuCacheMock.MockDownloader, [][]byte, error) {
	//load dummy data for download mock
	const menuBaseURL = "https://www.uksh.de/servicesternnord/Unser+Speisenangebot/Speisepl%C3%A4ne+L%C3%BCbeck/UKSH_Bistro+L%C3%BCbeck-p-346.html"
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
	downloadMock.EXPECT().Get(menuBaseURL).Return(dummySite, nil)
	downloadMock.EXPECT().Get(pdf1URL).Return(dummyPDF1, nil)
	downloadMock.EXPECT().Get(pdf2URL).Return(dummyPDF2, nil)

	return downloadMock, [][]byte{dummyPDF1, dummyPDF2}, nil
}

func TestExtractPDFsFromMenuSite(t *testing.T) {
	ctrl := gomock.NewController(t)

	downloadMock, expRes, err := createDownloaderMock(ctrl)
	if err != nil {
		t.Fatal(err)
	}

	resultPDFs, err := extractPDFsFromMenuSite(downloadMock)
	if err != nil {
		t.Errorf("Unexpected Error: %v\n", err)
	}
	if l := len(resultPDFs); l != 2 {
		t.Errorf("Expected 2 PDFs got %v\n", l)
	}

	if !reflect.DeepEqual(resultPDFs, expRes) {
		t.Errorf("Wrong PDFs returned\n")
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

	parseMock := parserMock.NewMockUKSHParser(ctrl)
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
