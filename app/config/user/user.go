package user

import (
	"github.com/unqnown/esctl/pkg/io"
	"log"
	"os"
	"strings"

	"github.com/unqnown/esctl/config"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:                   "user",
	Usage:                  "Shows current user.",
	Action:                 user,
	UseShortOptionHandling: true,
	Subcommands: []cli.Command{
		{
			Name:                   "add",
			Aliases:                []string{"assign"},
			Usage:                  "Set user to context.",
			ArgsUsage:              "name@password [-a]",
			Action:                 add,
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "assign, a",
					Usage: "Assign user to current context.",
				},
			},
		},
	},
}

func user(c *cli.Context) {
	conf, err := config.Open(c.GlobalString("config"))
	check.Fatalf(err, "open config: %v", err)

	usr, err := conf.User()
	check.Fatal(err)

	if usr.Nil {
		log.Printf("no user")
		return
	}

	log.Printf("%s", usr.Name)
}

func add(c *cli.Context) {
	args := c.Args()
	if !args.Present() {
		log.Fatal("user not specified")
	}

	conf, err := config.Open(c.GlobalString("config"))
	check.Fatalf(err, "open config: %v", err)

	usrpass := strings.Split(args.First(), "@")
	if len(usrpass) != 2 {
		log.Fatalf("incorrect user")
	}

	usr := config.User{
		Name:     usrpass[0],
		Password: usrpass[1],
	}

	conf.Users.Add(usr)

	var w io.Buffer

	w.Writef("user %q added\n", usr.Name)

	if c.Bool("assign") {
		ctx, err := conf.Ctx()
		check.Fatal(err)

		ctx.User = &usr.Name

		conf.Contexts[conf.Context] = ctx

		w.Writef("user %q assigned to %q\n", usr.Name, conf.Context)
	}

	err = conf.Save(c.GlobalString("config"))
	check.Fatalf(err, "save config: %v", err)

	_, _ = w.WriteTo(os.Stdout)
}
