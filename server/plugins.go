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
	"stash.kopano.io/kc/kapi/plugins"
)

func (s *Server) loadPlugins(enabledPlugins []string) error {
	var enabledPluginsMap map[string]bool
	if len(enabledPlugins) > 0 {
		enabledPluginsMap = make(map[string]bool)
		for _, id := range enabledPlugins {
			enabledPluginsMap[id] = true
		}
	}

	loadedPlugins := make(map[string]plugins.Plugin)
	for id, register := range plugins.Registered() {
		if enabledPluginsMap != nil {
			if !enabledPluginsMap[id] {
				// Skip plugin when not enabled.
				s.logger.WithField("plugin", id).Debugf("plugin not enabled, skipped: %s", id)
				continue
			}
		} else {
			// Auto enable plugin when none is explicitly enabled.
			enabledPlugins = append(enabledPlugins, id)
		}

		loadedPlugins[id] = register()
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
