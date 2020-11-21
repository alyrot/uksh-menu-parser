package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alyrot/menuToText/pkg/parser"
)

func main() {
	menuPDFPath := flag.String("menuPDF", "", "path to uksh menu pdf (mandatory)")

	flag.Parse()
	if *menuPDFPath == "" {
		fmt.Println("menuPDF is a mandatory argument")
		flag.PrintDefaults()
		os.Exit(0)
	}

	pdfBytes, err := ioutil.ReadFile(*menuPDFPath)
	if err != nil {
		fmt.Printf("Failed to read %v: %v", *menuPDFPath, err)
		os.Exit(1)
	}

	dishes, err := parser.PDFToDishes(pdfBytes)
	if err != nil {
		fmt.Printf("Failed to parse PDF: %v\n", err)
		os.Exit(1)
	}

	for i := range dishes {
		fmt.Printf("%v\n", dishes[i])
	}

}
