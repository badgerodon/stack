package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/badgerodon/stack/service"
	"github.com/badgerodon/stack/service/runner"
	"github.com/badgerodon/stack/storage"
	"github.com/codegangsta/cli"
	"github.com/kardianos/osext"
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

	if dst == "-" {
		source, err := storage.Get(snl)
		if err != nil {
			return err
		}
		defer source.Close()

		io.Copy(os.Stdout, source)
	} else {
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
	app.Usage = "a simple, cross-platform, open-source, pull-based deployment system"
	app.Version = "0.6"
	app.Author = "Caleb Doxsey"
	app.Email = "caleb@doxsey.net"
	app.Commands = []cli.Command{
		{
			Name:  "apply",
			Usage: "apply the configuration file",
			Action: func(c *cli.Context) {
				err := apply(c.Args().First())
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "auth",
			Usage: "generate credentials for services that need it",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					log.Fatalln("provider is required")
				}

				storage.Authenticate(c.Args()[0])
			},
		},
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
			Name:  "cat",
			Usage: "cat a file: cat <source>",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					log.Fatalln("source is required")
				}

				err := cp(c.Args()[0], "-")
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name:  "install",
			Usage: "install the stack as a service: install",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					log.Fatalln("config file location is required")
				}

				log.Println("installing using", serviceManager)

				// uninstall any existing stack service
				serviceManager.Uninstall("stack")

				exeName, err := osext.Executable()
				if err != nil {
					log.Fatalln(err)
				}

				dstName := filepath.Join(rootDir, "stack")

				log.Println("copying", exeName, "to", dstName)

				os.MkdirAll(rootDir, 0755)
				src, err := os.Open(exeName)
				if err != nil {
					log.Fatalln(err)
				}
				defer src.Close()

				dst, err := os.OpenFile(dstName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					log.Fatalln(err)
				}
				io.Copy(dst, src)
				dst.Close()

				err = serviceManager.Install(service.Service{
					Name:        "stack",
					Directory:   rootDir,
					Command:     []string{filepath.Join(rootDir, "stack"), "watch", c.Args().First()},
					Environment: map[string]string{},
				})
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
			Name:  "service-runner",
			Usage: "daemon started by `watch` that runs applications",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address",
					Usage: "address to connect to",
				},
				cli.StringFlag{
					Name:  "state-file",
					Usage: "file to store state in",
				},
			},
			Action: func(c *cli.Context) {
				runner.Run(c.String("address"), c.String("state-file"))
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
