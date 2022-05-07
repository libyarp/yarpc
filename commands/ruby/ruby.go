package ruby

import "github.com/urfave/cli"

var Action = cli.Command{
	Name:   "ruby",
	Usage:  "Compiles provided yarp files into Ruby sources",
	Action: nil,
}
