package main

import (
	"log"
	"strings"

	"github.com/cheggaaa/go-poppler"
)

func pdfMatchesPattern(filename string, pattern string) bool {
	doc, _ := poppler.Open(filename)
	defer doc.Close()
	nPages := doc.GetNPages()
	for pn := range nPages {
		page := doc.GetPage(pn)
		if strings.Contains(
			strings.ToLower(page.Text()),
			strings.ToLower(pattern)) {
			log.Println("Matching PDF page:", page.Text())
			page.Close()
			return true
		}
		page.Close()
	}
	return false
}
