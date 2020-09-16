package ctl

import (
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/unqnown/esctl/pkg/client"
	"github.com/urfave/cli"
)

type ActionFunc = func(*cli.Context)

type CommandFunc func(conf app.Config, cli *client.Client, ctx *cli.Context) error

func NewAction(cmd CommandFunc) ActionFunc {
	return func(c *cli.Context) {
		conf, err := app.Open(c.GlobalString("config"))
		check.Fatalf(err, "open config: %v", err)

		err = conf.SetContext(c.GlobalString("context"))
		check.Fatal(err)

		cst, usr, err := conf.Conn()
		check.Fatal(err)

		conn, err := client.New(cst, usr)
		check.Fatalf(err, "connect: %v", err)

		check.Fatal(cmd(conf, conn, c))
	}
}

func Call(cmd CommandFunc, conf app.Config, conn *client.Client) ActionFunc {
	return func(ctx *cli.Context) {
		check.Fatal(cmd(conf, conn, ctx))
	}
}
