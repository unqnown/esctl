package config

import (
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

func Default() Config {
	return Config{
		Users: map[string]User{},
		Clusters: map[string]Cluster{
			"localhost": {
				Servers: []string{"http://localhost:9200"},
			},
		},
		Contexts: map[string]Context{
			"default": {
				Cluster:  "localhost",
				Location: "default",
			},
		},
		Context: "default",
	}
}

type Location struct {
	Mappings string `yaml:"mappings"`
	Backups  string `yaml:"backups"`
	Queries  string `yaml:"queries"`
}

type Settings struct{}

type Config struct {
	Locations map[string]Location `yaml:"locations,omitempty"`
	Settings  Settings            `yaml:"settings,omitempty"`

	Users    Users              `yaml:"users,omitempty"`
	Clusters map[string]Cluster `yaml:"clusters"`

	Contexts map[string]Context `yaml:"contexts"`

	Context string `yaml:"context"`
}

func (conf Config) Location() (Location, error) {
	ctx, err := conf.Ctx()
	if err != nil {
		return Location{}, err
	}
	loc, found := conf.Locations[ctx.Location]
	if !found {
		return Location{}, errors.Errorf("location %q not found", ctx.Location)
	}
	return loc, nil
}

func (conf *Config) SetContext(ctx string) error {
	if ctx == "" {
		return nil
	}
	if _, exists := conf.Contexts[ctx]; !exists {
		return errors.Errorf("context %q not found", ctx)
	}
	conf.Context = ctx
	return nil
}

type Users map[string]User

func (u *Users) Add(usr User) {
	if *u == nil {
		*u = make(Users)
	}
	u.add(usr)
}

func (u Users) add(usr User) { u[usr.Name] = usr }

func (conf *Config) Validate() error {
	if _, err := conf.Ctx(); err != nil {
		return err
	}
	if _, err := conf.User(); err != nil {
		return err
	}
	if _, err := conf.Cluster(); err != nil {
		return err
	}
	if _, err := conf.Location(); err != nil {
		return err
	}
	return nil
}

func Open(name string) (conf Config, err error) { return conf, conf.Open(name) }

func (conf *Config) Open(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(conf); err != nil {
		return err
	}
	return conf.Validate()
}

func (conf Config) Save(path string) error {
	if err := conf.Validate(); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(conf)
}

func (conf Config) Ctx() (Context, error) {
	ctx, found := conf.Contexts[conf.Context]
	if !found {
		return Context{}, errors.Errorf("context %q not found", conf.Context)
	}
	return ctx, nil
}

func (conf Config) Cluster() (Cluster, error) {
	ctx, err := conf.Ctx()
	if err != nil {
		return Cluster{}, err
	}
	cst, found := conf.Clusters[ctx.Cluster]
	if !found {
		return Cluster{}, errors.Errorf("cluster %q not found", ctx.Cluster)
	}
	return cst, nil
}

func (conf Config) User() (User, error) {
	ctx, err := conf.Ctx()
	if err != nil {
		return User{}, err
	}
	if ctx.User == nil {
		return User{Nil: true}, nil
	}
	usr, found := conf.Users[*ctx.User]
	if !found {
		return User{}, errors.Errorf("user %q not found", *ctx.User)
	}
	return usr, nil
}

func (conf Config) Conn() (cst Cluster, usr User, err error) {
	if cst, err = conf.Cluster(); err != nil {
		return cst, usr, err
	}
	if usr, err = conf.User(); err != nil {
		return cst, usr, err
	}
	return cst, usr, nil
}

type Context struct {
	User     *string `yaml:"user,omitempty"`
	Cluster  string  `yaml:"cluster"`
	Location string  `yaml:"location"`
}

type Cluster struct {
	Servers  []string `yaml:"servers"`
	Settings Settings `yaml:"settings,omitempty"`
}

type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	// TODO(d.andriichuk): refactor user flow.
	Nil bool `yaml:"-"`
}
