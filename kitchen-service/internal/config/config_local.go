package config

import (
	"fmt"

	"github.com/w-k-s/konfig"
	"github.com/w-k-s/konfig/loader/klfile"
	"github.com/w-k-s/konfig/parser/kptoml"
)

type localConfigSource struct{}

func (l localConfigSource) Load(absolutePath string) (*Config, error) {
	configStore := konfig.New(konfig.DefaultConfig())
	configStore.RegisterLoaderWatcher(
		klfile.New(
			&klfile.Config{
				Files: []klfile.File{
					{
						Parser: kptoml.Parser,
						Path:   absolutePath,
					},
				},
			},
		),
	)

	if err := configStore.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config file from local path '%s'. Reason: %w", absolutePath, err)
	}

	return readValues(configStore)
}
