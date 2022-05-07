package main

import (
	"github.com/libyarp/yarpc/commands/golang"
	"github.com/libyarp/yarpc/commands/ruby"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.App{
		Name:        "yarpc",
		HelpName:    "yarpc",
		ArgsUsage:   "file [file...]",
		Version:     "0.0.1-dev",
		Description: "Yarp Compiler",
		Commands: []cli.Command{
			golang.Action,
			ruby.Action,
		},
		Flags: []cli.Flag{},
		Authors: []cli.Author{
			{
				Name:  "Victor \"Vito\" Gama",
				Email: "hey@vito.io",
			},
		},
		Copyright: "(c) 2022 - libyarp authors",
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
