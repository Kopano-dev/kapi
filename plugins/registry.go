/*
 * Copyright 2020 Kopano and its licensors
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

package plugins

var registry map[string]func() Plugin

func init() {
	registry = make(map[string]func() Plugin)
}

// RegisterV1 is the function where V1 plugins can register themselves. A plugin
// needs to be registered so it can be found by consumers.
func RegisterV1(name string, registerFunc RegisterPluginV1) error {
	registry[name] = func() Plugin {
		return registerFunc()
	}
	return nil
}

// Registered returns the register function for all registered
// plugins by name.
func Registered() map[string]func() Plugin {
	plugins := make(map[string]func() Plugin)
	for name, registerFunc := range registry {
		plugins[name] = registerFunc
	}

	return plugins
}
