package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func getMozillaFiles() []string {
	files := make([]string, 0)

	user, _ := user.Current()
	baseDir := filepath.Join(user.HomeDir, ".mozilla/firefox")

	items, err := os.ReadDir(baseDir)
	if err == nil {
		for _, f := range items {
			if !f.IsDir() {
				continue
			}

			if !strings.Contains(f.Name(), "default") {
				continue
			}

			placesFile := filepath.Join(baseDir, f.Name(), "places.sqlite")
			_, err := os.Stat(placesFile)
			if err != nil {
				log.Println(err)
				continue
			}

			files = append(files, placesFile)
		}
	} else {
		log.Println(err)
	}

	return files
}

/*
func main() {
	for _, file := range getMozillaFiles() {
		fmt.Println(file)
	}
}
*/
