package main

import (
	"fmt"
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
			fmt.Println("Fragment:", page.Text())
			page.Close()
			return true
		}
		page.Close()
	}
	return false
}
