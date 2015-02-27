package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/badgerodon/stack/storage"
	"github.com/codegangsta/cli"
)

func cp(src, dst string) error {
	if strings.HasSuffix(dst, "/") && !strings.HasSuffix(src, "/") {
		n := src
		if strings.Contains(n, "/") {
			n = n[strings.LastIndex(n, "/")+1:]
		}
		dst = filepath.Join(dst, n)
	}

	snl, err := storage.ParseLocation(src)
	if err != nil {
		return err
	}

	dnl, err := storage.ParseLocation(dst)
	if err != nil {
		return err
	}

	source, err := storage.Get(snl)
	if err != nil {
		return err
	}
	defer source.Close()

	err = storage.Put(dnl, source)
	if err != nil {
		return err
	}
	return nil
}

func ls(dir string) error {
	loc, err := storage.ParseLocation(dir)
	if err != nil {
		return err
	}
	names, err := storage.List(loc)
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func rm(src string) error {
	loc, err := storage.ParseLocation(src)
	if err != nil {
		return err
	}
	return storage.Delete(loc)
}

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

				err := cp(c.Args()[0], c.Args()[1])
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "ls",
			Usage: "list a directory",
			Action: func(c *cli.Context) {
				err := ls(c.Args().First())
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "rm",
			Usage: "remove a file",
			Action: func(c *cli.Context) {
				err := rm(c.Args().First())
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "install",
			Usage: "install the stack service",
			Action: func(c *cli.Context) {
				err := install(c.Args().First())
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "watch",
			Usage: "watch a config file",
			Action: func(c *cli.Context) {
				err := watch(c.Args().First())
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
	}
	app.Run(os.Args)
}
