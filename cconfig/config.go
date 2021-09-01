package cconfig

import (
	"errors"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/gocopper/copper/cerrors"
	"github.com/imdario/mergo"
	"github.com/pelletier/go-toml"
)

type (
	// Env defines the various environments the app can be configured for.
	// Examples of Env include dev, test, staging, or prod.
	Env string

	// Dir defines the directory where config file(s) live.
	Dir string

	// ProjectDir defines the project directory. This variable can be used in the config file with {{ .ProjectDir }}.
	// It is set by passing a -project flag to the app binary.
	ProjectDir string
)

// Config provides methods to read app config.
type Config interface {
	Load(key string, dest interface{}) error
}

const (
	baseTomlConfigFileName  = "base.toml"
	localTomlConfigFileName = "local.toml"
	tomlExt                 = ".toml"
)

// New provides an implementation of Config that reads config files in the
// dir. By default, it reads from base.toml and can be overridden by a file
// corresponding to the env. For 'test' env, the file should be test.toml.
func New(dir Dir, projectDir ProjectDir, env Env) (Config, error) {
	var (
		baseConfigPath  = path.Join(string(dir), baseTomlConfigFileName)
		envConfigPath   = path.Join(string(dir), strings.ToLower(string(env))+tomlExt)
		localConfigPath = path.Join(string(dir), localTomlConfigFileName)
	)

	baseTree, err := loadTOMLTemplate(baseConfigPath, projectDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, cerrors.New(err, "failed to load base config file", map[string]interface{}{
			"path": baseConfigPath,
		})
	}

	envTree, err := loadTOMLTemplate(envConfigPath, projectDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, cerrors.New(err, "failed to load env config file", map[string]interface{}{
			"env":  env,
			"path": envConfigPath,
		})
	}

	localTree, err := loadTOMLTemplate(localConfigPath, projectDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, cerrors.New(err, "failed to load local config file", map[string]interface{}{
			"path": localConfigPath,
		})
	}

	return &config{
		base:       baseTree,
		env:        envTree,
		local:      localTree,
		currentEnv: env,
	}, nil
}

type config struct {
	base       *toml.Tree
	env        *toml.Tree
	local      *toml.Tree
	currentEnv Env
}

func (c *config) Load(key string, dest interface{}) error {
	var (
		base  = &toml.Tree{}
		env   = &toml.Tree{}
		local = &toml.Tree{}
	)

	if c.base != nil && c.base.Has(key) {
		base = c.base.Get(key).(*toml.Tree)
	}

	if c.env != nil && c.env.Has(key) {
		env = c.env.Get(key).(*toml.Tree)
	}

	if c.local != nil && c.local.Has(key) {
		local = c.local.Get(key).(*toml.Tree)
	}

	err := toml.Unmarshal([]byte(""), dest)
	if err != nil {
		return cerrors.New(err, "failed to load config defaults", nil)
	}

	err = c.loadWithNoDefaults(base, dest)
	if err != nil {
		return cerrors.New(err, "failed to load base config", nil)
	}

	err = c.loadWithNoDefaults(env, dest)
	if err != nil {
		return cerrors.New(err, "failed to load env config", nil)
	}

	err = c.loadWithNoDefaults(local, dest)
	if err != nil {
		return cerrors.New(err, "failed to load local config", nil)
	}

	return nil
}

func (c *config) loadWithNoDefaults(t *toml.Tree, dest interface{}) error {
	defaults := reflect.New(reflect.TypeOf(dest).Elem()).Interface()

	err := toml.Unmarshal([]byte(""), defaults)
	if err != nil {
		return cerrors.New(err, "failed to load defaults", nil)
	}

	vals := reflect.New(reflect.TypeOf(dest).Elem()).Interface()

	err = t.Unmarshal(vals)
	if err != nil {
		return cerrors.New(err, "failed to load config values with defaults", nil)
	}

	err = mergo.Merge(vals, defaults, mergo.WithTransformers(c))
	if err != nil {
		return cerrors.New(err, "failed to remove default values from config", nil)
	}

	err = mergo.Merge(dest, vals, mergo.WithOverride)
	if err != nil {
		return cerrors.New(err, "failed to merge config vals with dest", nil)
	}

	return nil
}

func (c *config) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Struct {
		return nil
	}

	return func(dst, src reflect.Value) error {
		if dst.Interface() == src.Interface() {
			dst.Set(reflect.Zero(dst.Type()))
		}

		return nil
	}
}