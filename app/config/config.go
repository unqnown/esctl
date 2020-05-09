package config

import (
	"github.com/unqnown/esctl/app/config/context"
	"github.com/unqnown/esctl/app/config/user"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "config",
	Usage:                  "Config management.",
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		context.Command,
		user.Command,
	},
}
