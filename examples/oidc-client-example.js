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

'use strict';

window.app = new Vue({
	el: '#app',
	data: {
		error: null,
		userManager: null,
		user: null,
		expires_in: 0,
		initialized: false,
		iss: '',
		requestResponse: ''
	},
	components: {
	},
	created: function() {
		console.info('welcome oidc-client-eample');

		const queryValues = parseParams(location.search.substr(1));
		console.log('URL query values on load', queryValues);

		let iss = queryValues.iss;
		if (!iss) {
			iss = localStorage.getItem('oidc-client-example.iss');
		}

		this.iss = iss;

		setInterval(() => {
			if (this.user) {
				const expires_in = this.user.expires_in;
				this.$nextTick(() => {
					this.expires_in = expires_in;
				});
			}
		}, 500);

		if (this.iss) {
			this.$nextTick(() => {
				this.initialize().then(() => {
					this.getUser();
				})
			});
		}
	},
	watch: {
		iss: val => {
			if (val) {
					localStorage.setItem('oidc-client-example.iss', val);
			} else {
					localStorage.removeItem('oidc-client-example.iss');
			}
		},
		user: user => {
			console.log('user updated', user);
			if (user) {
				this.expires_in = user.expires_in;
			}
		}
	},
	methods: {
		initialize: function() {
			const callbackURI = window.location.href.split('#')[0] + '#callback';

			const mgr = new Oidc.UserManager({
				authority: this.iss,
				client_id: 'oidc-client-example',
				redirect_uri: callbackURI,
				response_type: 'id_token token',
				scope: 'openid profile email',
				loadUserInfo: true
			});
			mgr.events.addAccessTokenExpiring(() => {
				console.log('token expiring');
			});
			mgr.events.addUserLoaded(() => {
				console.log('user loaded');
			});

			this.userManager = mgr;

			if (window.location.href.indexOf(callbackURI) === 0) {
				window.location.hash = window.location.hash.substr(9);
				return this.completeAuthentication().then(() => {
					console.log('completed authentication', this.user);
				}).catch((err) => {
					console.log('failed to complete authentication', err);
				}).then(() => {
					window.location.hash = '';
					this.initialized = true;
				});
			} else {
				this.initialized = true;
				return Promise.resolve(null);
			}
		},

		startAuthentication: async function() {
			return this.userManager.signinRedirect();
		},

		completeAuthentication: async function() {
			return this.userManager.signinRedirectCallback().then(user => {
				this.user = user;

				return user;
			});
		},

		isLoggedIn: function() {
			this.user != null && !this.user.expired;
		},

		getUser: async function() {
			return this.userManager.getUser().then((user) => {
				this.user = user;

				if (user) {
					Vue.http.headers.common['Authorization'] = user.token_type + ' ' + user.access_token;
				}
				return user;
			}).catch((err) => {
				console.error('failed to get user', err);
				this.user = null;

				return null;
			});
		},

		gcGet: async function(endpoint) {
			const prefix = '/api/gc/v0';

			this.requestResponse = '';
			return this.$http.get(prefix + endpoint).then(response => {
				// Whoohoo success.
				this.requestResponse = response.body;
				return response.body;
			});
		},
		gcFetchFolders: async function() {
			return this.gcGet('/folders').then(response => {
				return response;
			});
		},
		gcFetchUsers: async function() {
			return this.gcGet('/users').then(response => {
				return response;
			});
		},
		gcFetchStores: async function() {
			return this.gcGet('/stores').then(response => {
				return response;
			});
		}
	}
});
