package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/unqnown/semver"
)

var (
	ErrContextNotFound = errors.New("context: not found")
	ErrClusterNotFound = errors.New("cluster: not found")
	ErrUserNotFound    = errors.New("user: not found")
)

const (
	BackupDir  = ".backup"
	MappingDir = ".mapping"
	QueryDir   = ".query"
)

func NewConfig(ver semver.Version, home string) Config {
	return Config{
		Version: ver,
		Home:    home,
		Users:   map[string]User{},
		Clusters: map[string]Cluster{
			"localhost": {
				Servers: []string{"http://localhost:9200"},
			},
		},
		Contexts: map[string]Context{
			"default": {
				Cluster: "localhost",
			},
		},
		Context: "default",
	}
}

type Settings struct{}

type Config struct {
	Version semver.Version `yaml:"version"`

	Home     string   `yaml:"home"`
	Settings Settings `yaml:"settings,omitempty"`

	Users    Users              `yaml:"users,omitempty"`
	Clusters map[string]Cluster `yaml:"clusters"`

	Contexts map[string]Context `yaml:"contexts"`

	Context string `yaml:"context"`
}

func (conf *Config) SetContext(ctx string) error {
	if ctx == "" {
		return nil
	}
	if _, exists := conf.Contexts[ctx]; !exists {
		return fmt.Errorf("set %q context: %w", ctx, ErrContextNotFound)
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
		return Context{}, fmt.Errorf("get %q context: %w", conf.Context, ErrContextNotFound)
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
		return Cluster{}, fmt.Errorf("get %q cluster: %w", ctx.Cluster, ErrClusterNotFound)
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
		return User{}, fmt.Errorf("get %q user: %w", *ctx.User, ErrUserNotFound)
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
	User    *string `yaml:"user,omitempty"`
	Cluster string  `yaml:"cluster"`
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
