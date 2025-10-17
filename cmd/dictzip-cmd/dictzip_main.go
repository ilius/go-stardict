package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ilius/go-stardict/v2/dictzip"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <file.dz>", os.Args[0])
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader, err := dictzip.NewReader(file)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	const chunkSize = 10
	b := make([]byte, chunkSize)
	pos := int64(0)

	for {
		n, err := reader.ReadAt(b, pos)

		if n > 0 {
			fmt.Print(string(b[:n]))
			pos += int64(n)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading at pos=%v: %v", pos, err)
		}
	}
	fmt.Printf("\n--- done (read %d bytes) ---\n", pos)
}
