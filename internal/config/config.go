package config

import (
	"errors"
	"fmt"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path"
	"strconv"
	"strings"
	"time"
)

// ParsedConfig is holder for resolved configuration
type ParsedConfig struct {
	CA        CertConfig `yaml:"ca"`
	Signables []CertConfig
	BaseDir   string
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

	m := make(map[string]CertConfig)
	err = yaml.Unmarshal(confBytes, &m)
	if err != nil {
		return nil, err
	}

	baseDir := path.Dir(yamlPath)

	conf := ParsedConfig{}
	defaults, ok := m["default"]
	if !ok {
		defaults = CertConfig{}
	}

	if c, ok := m["ca"]; ok {
		certConf, err := resolveCertConf("ca", c, defaults)
		if !strings.HasPrefix(certConf.Path, "/") {
			certConf.Path = fmt.Sprintf("%s/%s", baseDir, certConf.Path)
		}
		if err != nil {
			return nil, err
		}
		conf.CA = *certConf
	}

	for k, v := range m {
		if k == "default" || k == "ca" {
			continue
		}
		certConf, err := resolveCertConf(k, v, defaults)

		// Path defined will be relative to config file base directory.
		// Unelss it is defined as absolute path.
		if !strings.HasPrefix(certConf.Path, "/") {
			certConf.Path = fmt.Sprintf("%s/%s", baseDir, certConf.Path)
		}

		if err != nil {
			return nil, err
		}
		conf.Signables = append(conf.Signables, *certConf)
	}

	if debug {
		log.Println(conf)
	}
	return &conf, nil
}

func resolveCertConf(name string, conf, defaults CertConfig) (*CertConfig, error) {
	baseConf := defaults
	baseConf.Name = name
	if err := mergo.MergeWithOverwrite(&baseConf, conf); err != nil {
		return nil, err
	}

	if conf.TTL.ttl == "" {
		baseConf.TTL = defaults.TTL
	}

	return &baseConf, nil
}
