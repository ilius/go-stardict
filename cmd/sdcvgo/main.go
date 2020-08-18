package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	stardict "github.com/ilius/go-stardict"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dicDir := path.Join(homeDir, ".stardict", "dic")
	dics, err := stardict.Open(dicDir)
	if err != nil {
		panic(err)
	}
	for _, word := range os.Args[1:] {
		for dicI, dic := range dics {
			if dicI > 0 {
				fmt.Printf("\n")
			}
			transList := dic.Translate(word)
			if len(transList) > 0 {
				fmt.Printf("--> query %#v from %s\n", word, dic.GetBookName())
			}
			for transI, trans := range transList {
				if transI > 0 {
					fmt.Printf("----------\n")
				}
				for _, part := range trans.Parts {
					fmt.Printf("%v\n", strings.TrimSpace(string(part.Data)))
				}
				fmt.Printf("\n")
			}
		}
	}
}
