package plugin

import (
	"plugin"

	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/model"
	"github.com/pkg/errors"
)

var loadedGoPlugin map[string]*plugin.Plugin = make(map[string]*plugin.Plugin)

type Interface interface {
	RegisterSlashCommand() []model.Command
}

type Plugin struct {
	Interface

	Name string
}

func Load(cfg []config.PluginConfig) ([]Plugin, error) {
	plugins := make([]Plugin, len(cfg))

	for i, pluginCfg := range cfg {
		// if we already loaded the go plugin, reuse it
		p, ok := loadedGoPlugin[pluginCfg.File]
		if !ok {
			var err error
			p, err = plugin.Open(pluginCfg.File)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to open plugin %s", pluginCfg.Name)
			}
			loadedGoPlugin[pluginCfg.File] = p
		}

		lookupNew, err := p.Lookup("New")
		if err != nil {
			return nil, errors.Wrapf(err, "failed to lookup New in plugin %s", pluginCfg.Name)
		}
		newFn, ok := lookupNew.(func(config.RawMessage) (Interface, error))
		if !ok {
			return nil, errors.Errorf("failed to find New in plugin %s", pluginCfg.Name)
		}

		createdPlugin, err := newFn(pluginCfg.Config)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create plugin %s", pluginCfg.Name)
		}

		plugins[i] = Plugin{
			Interface: createdPlugin,
			Name:      pluginCfg.Name,
		}
	}

	return plugins, nil
}
