/*
 * Copyright 2017 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/sirupsen/logrus"

	"stash.kopano.io/kc/kapi/plugins"
)

func (s *Server) loadPlugins(enabledPlugins []string) error {
	if s.pluginsPath == "" {
		return nil
	}

	if fp, err := os.Stat(s.pluginsPath); err != nil || !fp.IsDir() {
		return fmt.Errorf("plugin directory does not exist or is not directory")
	}

	var enabledPluginsMap map[string]bool
	if len(enabledPlugins) > 0 {
		enabledPluginsMap = make(map[string]bool)
		for _, id := range enabledPlugins {
			enabledPluginsMap[id] = true
		}
	}

	loadedPlugins := make(map[string]plugins.Plugin)
	err := filepath.Walk(s.pluginsPath, func(path string, fp os.FileInfo, _ error) error {
		id, p := s.loadPlugin(path, fp)
		if id == "" || p == nil {
			return nil
		}

		if enabledPluginsMap != nil {
			if !enabledPluginsMap[id] {
				// Skip plugin when not enabled.
				s.logger.WithField("plugin", id).Debugf("plugin not enabled, skipped: %s", path)
				return nil
			}
		} else {
			// Auto enable plugin when none is explicitly enabled.
			enabledPlugins = append(enabledPlugins, id)
		}

		loadedPlugins[id] = p
		return nil
	})
	if err != nil {
		return err
	}

	for _, id := range enabledPlugins {
		if p, ok := loadedPlugins[id]; ok {
			s.plugins = append(s.plugins, p)
			s.logger.WithField("plugin", id).Infoln("plugin registered")
		} else {
			s.logger.WithField("plugin", id).Warnln("plugin not found")
		}
	}

	return nil
}

func (s *Server) loadPlugin(path string, fp os.FileInfo) (string, plugins.Plugin) {
	if fp.IsDir() {
		return "", nil
	}

	// NOTE(longsleep): Allow try only files which contain .so. This includes
	// for example `plugin.so` and also `plugin.so.1`.
	if !strings.Contains(fp.Name(), ".so") {
		return "", nil
	}

	p, err := plugin.Open(path)
	if err != nil {
		s.logger.WithError(err).Debugf("plugin invalid: %s", path)
		return "", nil
	}

	if registerLookup, err := p.Lookup("Register"); err == nil {
		if register, ok := registerLookup.(*plugins.RegisterPluginV1); ok {
			p := (*register)()

			info := p.Info()
			if info.ID == "" {
				s.logger.Warnf("plugin without ID, skipped: %s", path)
				return "", nil
			}
			s.logger.WithFields(logrus.Fields{
				"plugin":  info.ID,
				"version": info.Version,
				"build":   info.BuildDate,
			}).Infof("plugin loaded: %s", path)
			return info.ID, p
		} else {
			s.logger.Warnf("plugin type %#v unknown: %s", registerLookup, path)
		}
	} else {
		s.logger.WithError(err).Debugf("plugin implementation invalid: %s", path)
	}

	return "", nil
}
