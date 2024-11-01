package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	stardict "github.com/ilius/go-stardict/v2"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dicDir := path.Join(homeDir, ".stardict", "dic")
	dics, err := stardict.Open(
		[]string{dicDir},
		map[string]int{},
	)
	if err != nil {
		panic(err)
	}
	for _, word := range os.Args[1:] {
		for dicI, dic := range dics {
			if dicI > 0 {
				fmt.Printf("\n")
			}
			results := dic.SearchFuzzy(word, 8, 5*time.Second)
			if len(results) > 0 {
				fmt.Printf("--> query %#v from %s\n", word, dic.DictName())
			}
			for index, result := range results {
				if index > 0 {
					fmt.Printf("----------\n")
				}
				for _, item := range result.Items() {
					fmt.Printf("%v\n", strings.TrimSpace(string(item.Data)))
				}
				fmt.Printf("\n")
			}
		}
	}
}
