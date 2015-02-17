package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatalln("expected mode")
	}

	switch os.Args[1] {
	case "watch":
		if len(os.Args) < 3 {
			log.Fatalln("expected endpoint")
		}
		Watch(os.Args[2])
	default:
		log.Fatalln("unknown mode")
	}
}
