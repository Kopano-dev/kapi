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

	"stash.kopano.io/kc/kopano-api/plugins"
)

func (s *Server) loadPlugins() error {
	if s.pluginsPath == "" {
		return nil
	}

	if fp, err := os.Stat(s.pluginsPath); err != nil || !fp.IsDir() {
		return fmt.Errorf("plugin directory does not exist or is not directory")
	}

	err := filepath.Walk(s.pluginsPath, s.loadPlugin)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) loadPlugin(path string, fp os.FileInfo, _ error) error {
	if fp.IsDir() {
		return nil
	}

	// NOTE(longsleep): Allow try only files which contain .so. This includes
	// for example `plugin.so` and also `plugin.so.1`.
	if !strings.Contains(fp.Name(), ".so") {
		return nil
	}

	p, err := plugin.Open(path)
	if err != nil {
		s.logger.WithError(err).Debugf("invalid plugin: %s", path)
		return nil
	}

	if registerLookup, err := p.Lookup("Register"); err == nil {
		if register, ok := registerLookup.(*plugins.RegisterPluginV1); ok {
			p := (*register)()

			info := p.Info()
			if info.ID == "" {
				s.logger.Warnf("skipping plugin without ID: %s", path)
				return nil
			}
			if s.enabledPlugins != nil && s.enabledPlugins[info.ID] != true {
				// Skip plugin when not enabled.
				s.logger.WithField("plugin", info.ID).Debugf("skipping not enabled plugin: %s", path)
				return nil
			}

			s.plugins = append(s.plugins, p)
			s.logger.WithFields(logrus.Fields{
				"plugin":  info.ID,
				"version": info.Version,
				"build":   info.BuildDate,
			}).Infof("registered plugin: %s", path)
		} else {
			s.logger.Warnf("unknown plugin type %#v: %s", registerLookup, path)
		}
	} else {
		s.logger.WithError(err).Debugf("invalid plugin implementation: %s", path)
	}

	return nil
}
