package app

import (
	"log"
	"os"

	"github.com/unqnown/esctl/app/alias"
	"github.com/unqnown/esctl/app/close"
	"github.com/unqnown/esctl/app/config"
	"github.com/unqnown/esctl/app/create"
	"github.com/unqnown/esctl/app/delete"
	"github.com/unqnown/esctl/app/dump"
	"github.com/unqnown/esctl/app/get"
	"github.com/unqnown/esctl/app/open"
	"github.com/unqnown/esctl/app/reindex"
	"github.com/unqnown/esctl/app/replace"
	"github.com/unqnown/esctl/app/restore"
	"github.com/unqnown/esctl/app/top"
	"github.com/unqnown/esctl/app/vacuum"
	"github.com/unqnown/esctl/pkg/check"
	"github.com/urfave/cli"
	"path/filepath"

	appconfig "github.com/unqnown/esctl/config"
)

func Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	app := cli.NewApp()

	app.Name = "esctl"
	app.Version = "v0.1.0"
	app.Usage = "Elasticsearch cluster managing tool."
	app.Description = "To start using esctl immediately run `esctl init`."
	app.UseShortOptionHandling = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, conf",
			Usage:  "Path to config `FILE`",
			Value:  filepath.Join(home, ".esctl/config.yaml"),
			EnvVar: "ESCTLCONFIG",
		},
		cli.StringFlag{
			Name:  "context, ctx",
			Usage: "Context to apply",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:                   "init",
			Usage:                  "Initialize or reinitialize esctl.",
			Description:            "Creates default config file.",
			Action:                 _init,
			UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force esctl reinitialization",
				},
			},
		},
		config.Command,
		get.Command,
		create.Command,
		top.Command,
		alias.Command,
		delete.Command,
		dump.Command,
		restore.Command,
		reindex.Command,
		vacuum.Command,
		replace.Command,
		open.Command,
		close.Command,
	}

	return app.Run(os.Args)
}

func _init(ctx *cli.Context) {
	path := ctx.GlobalString("config")

	switch _, err := os.Stat(path); {
	case err == nil:
		if !ctx.Bool("force") {
			log.Fatalf("config already exists")
		}
	case os.IsNotExist(err):
		// go ahead
	default:
		log.Fatalf("check config exists")
	}

	dir, _ := filepath.Split(path)

	err := os.MkdirAll(dir, os.ModePerm)
	check.Fatal(err)

	conf := appconfig.Default()
	conf.Locations = map[string]appconfig.Location{
		"default": {
			Mappings: filepath.Join(dir, ".mappings"),
			Backups:  filepath.Join(dir, ".backups"),
			Queries:  filepath.Join(dir, ".queries"),
		},
	}

	err = conf.Save(path)
	check.Fatalf(err, "init config: %v", err)

	log.Printf("%q created", path)
}


