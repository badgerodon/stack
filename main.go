package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/badgerodon/stack/storage"
	"github.com/codegangsta/cli"
)

func main() {
	log.SetFlags(0)

	app := cli.NewApp()
	app.Name = "stack"
	app.Commands = []cli.Command{
		{
			Name:  "cp",
			Usage: "copy a file: cp <source> <dest>",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 2 {
					log.Fatalln("source and destination arguments are required")
				}

				sn, dn := c.Args()[0], c.Args()[1]

				if strings.HasSuffix(dn, "/") && !strings.HasSuffix(sn, "/") {
					n := sn
					if strings.Contains(n, "/") {
						n = n[strings.LastIndex(n, "/")+1:]
					}
					dn = filepath.Join(dn, n)
				}

				source, err := storage.Get(sn)
				if err != nil {
					log.Fatalln(err)
				}
				defer source.Close()

				err = storage.Put(dn, source)
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "ls",
			Usage: "list a directory",
			Action: func(c *cli.Context) {
				names, err := storage.List(c.Args().First())
				if err != nil {
					log.Fatalln(err)
				}
				for _, name := range names {
					log.Println(name)
				}
			},
		},
	}
	app.Run(os.Args)
}
