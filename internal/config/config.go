package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"log"
	"strconv"
	"strings"
	"time"
)

type defaultHolder struct {
	Default CertConfig
}

// ParsedConfig is holder for resolved configuration
type ParsedConfig struct {
	CA       CertConfig `yaml:"ca"`
	Route    CertConfig `yaml:"route"`
	Server   CertConfig `yaml:"server"`
	Client   CertConfig `yaml:"client"`
	Accounts CertConfig `yaml:"accounts"`
}

// CertConfig config holder for single key/certificate properties
type CertConfig struct {
	Name      string   `yaml:"name"`
	Path      string   `yaml:"path"`
	TTL       TTL      `yaml:"ttl"`
	KeyLength int      `yaml:"key_length"`
	DNS       []string `yaml:"dns"`
	Subject   CertSubject
}

// CertSubject certificate subject
type CertSubject struct {
	CN      string `yaml:"cn"`
	Org     string `yaml:"org"`
	Country string `yaml:"country"`
}

// TTL holds certificate expiration data and provides expiration timestamp
// when cofiguration YAML is unmarshalled
type TTL struct {
	ttl        string
	Expiration time.Time
}

// UnmarshalYAML provides custom unmarshalling for setting the expiration
func (t *TTL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	t.ttl = s
	f := strings.Fields(s)
	val, err := strconv.Atoi(f[0])
	if err != nil {
		return err
	}
	unit := f[1]

	exp := time.Now()
	switch unit {
	case "years", "year":
		t.Expiration = exp.AddDate(val, 0, 0)
	case "months", "month":
		t.Expiration = exp.AddDate(0, val, 0)
	case "days", "day":
		t.Expiration = exp.AddDate(0, 0, val)
	case "hours", "hour":
		t.Expiration = exp.Add(time.Duration(int(time.Hour) * val))
	case "minutes", "minute":
		t.Expiration = exp.Add(time.Duration(int(time.Minute) * val))
	case "seconds", "second":
		t.Expiration = exp.Add(time.Duration(int(time.Second) * val))
	default:
		return errors.New("Unsuported TTL unit: " + unit)
	}

	if err != nil {
		return err
	}
	return nil
}

func (t *TTL) String() string {
	return fmt.Sprintf("{expiration: %s}", t.Expiration)
}

// ParseConfig parses the config
func ParseConfig(yamlPath string, debug bool) (*ParsedConfig, error) {
	confBytes, err := ioutil.ReadFile(yamlPath)

	// Read the defaults section and set it as the based for other sections
	defs := &defaultHolder{}
	err = yaml.Unmarshal(confBytes, &defs)
	if err != nil {
		return nil, err
	}

	conf := &ParsedConfig{
		CA:       defs.Default,
		Route:    defs.Default,
		Server:   defs.Default,
		Client:   defs.Default,
		Accounts: defs.Default,
	}
	conf.CA.Name = "ca"
	conf.Route.Name = "route"
	conf.Server.Name = "server"
	conf.Client.Name = "client"
	conf.Accounts.Name = "accounts"

	err = yaml.Unmarshal(confBytes, &conf)
	if err != nil {
		return nil, err
	}

	if debug {
		log.Println(*conf)
	}
	return conf, nil
}
