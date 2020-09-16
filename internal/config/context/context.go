package context

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "context",
	Usage:                  "Shows current context.",
	Action:                 context,
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		{
			Name:                   "set",
			Aliases:                []string{"use"},
			Usage:                  "Set current context.",
			ArgsUsage:              "context",
			Action:                 set,
			UseShortOptionHandling: true,
		},
		{
			Name:                   "delete",
			Usage:                  "Delete context.",
			ArgsUsage:              "context...",
			Action:                 _delete,
			UseShortOptionHandling: true,
		},
		{
			Name:                   "add",
			Aliases:                []string{"new"},
			Usage:                  "Add new context.",
			ArgsUsage:              "context --user user --cluster cluster [-u]",
			Action:                 add,
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "use, u",
					Usage: "Switch to new context.",
				},
				cli.StringFlag{
					Name:  "user",
					Usage: "User",
				},
				cli.StringFlag{
					Name:     "cluster",
					Required: true,
					Usage:    "Cluster",
				},
			},
		},
	},
}

func context(c *cli.Context) {
	conf, err := app.Open(c.GlobalString("config"))
	check.Fatalf(err, "open config: %v", err)
	log.Printf("%s", conf.Context)
}

func add(c *cli.Context) {
	args := c.Args()
	if !args.Present() {
		log.Fatal("context not specified")
	}

	conf, err := app.Open(c.GlobalString("config"))
	check.Fatalf(err, "open config: %v", err)

	ctx := args.First()

	context, exists := conf.Contexts[ctx]

	usr := c.String("user")
	context.User = &usr
	context.Cluster = c.String("cluster")

	var w bytes.Buffer

	conf.Contexts[ctx] = context
	if exists {
		fmt.Fprintf(&w, "context %q updated\n", ctx)
	} else {
		context.Location = "default"
		fmt.Fprintf(&w, "context %q added\n", ctx)
	}

	if c.Bool("use") {
		conf.Context = ctx
		fmt.Fprintf(&w, "switched to context %q\n", conf.Context)
	}

	err = conf.Save(c.GlobalString("config"))
	check.Fatalf(err, "save config: %v", err)

	_, _ = w.WriteTo(os.Stdout)
}

func set(c *cli.Context) {
	args := c.Args()
	if !args.Present() {
		log.Fatal("context not specified")
	}

	conf, err := app.Open(c.GlobalString("config"))
	check.Fatalf(err, "open config: %v", err)

	ctx := args.First()

	if _, found := conf.Contexts[ctx]; !found {
		log.Fatalf("context %s not found", ctx)
	}

	conf.Context = ctx

	err = conf.Save(c.GlobalString("config"))
	check.Fatalf(err, "save config: %v", err)

	log.Printf("switched to context %q", conf.Context)
}

func _delete(c *cli.Context) {
	args := c.Args()
	if !args.Present() {
		log.Fatal("context not specified")
	}

	conf, err := app.Open(c.GlobalString("config"))
	check.Fatalf(err, "open config: %v", err)

	var w bytes.Buffer

	for _, ctx := range args {
		if ctx == conf.Context {
			log.Fatalf("%q is current context and can't be deleted", ctx)
		}
		delete(conf.Contexts, ctx)
		fmt.Fprintf(&w, "%q deleted\n", ctx)
	}

	err = conf.Save(c.GlobalString("config"))
	check.Fatalf(err, "save config: %v", err)

	_, _ = w.WriteTo(os.Stdout)
}
