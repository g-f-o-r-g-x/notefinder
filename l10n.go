package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"fyne.io/fyne/v2/lang"
)

var (
	translationFile  = "translation.json"
	makeLocalization = os.Getenv("NF_MAKE_L10N") != ""
	mu               sync.Mutex
)

func l10n(in string) string {
	localized := lang.Localize(in)

	if makeLocalization {
		mu.Lock()
		defer mu.Unlock()

		translations := make(map[string]string)
		data, err := os.ReadFile(translationFile)
		if err == nil {
			_ = json.Unmarshal(data, &translations)
		}

		if _, exists := translations[in]; !exists {
			translations[in] = ""
			newData, err := json.MarshalIndent(translations, "", "  ")
			if err == nil {
				_ = os.WriteFile(translationFile, newData, 0644)
			} else {
				fmt.Println("Error marshaling translations:", err)
			}
		}
	}
	return localized
}
