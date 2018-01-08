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

		apiPrefix: '/api/gc/v0',

		requestTab: '',
		requestKey: null,
		requestEndpoint: '',
		requestResponse: '',
		requestResponseHeaders: null,
		requestNextLink: null,

		requestMode: {
			default: true
		},
		requestStatus: null,
		requestResults: {},

		responseTab: '',
		responseMode: {
			default: true
		},

		createStatus: null
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

		Oidc.Log.logger = console;

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
		},
		requestTab: function(val) {
			const r = {};
			switch (val) {
			case undefined:
			case '':
				r.default = true;
				break;
			default:
				r[val] = true;
				break;
			}
			this.requestMode = r;
		},
		responseTab: function(val) {
			const r = {};
			switch (val) {
			case undefined:
			case '':
				r.default = true;
				break;
			default:
				r[val] = true;
				break;
			}
			this.responseMode = r;
		}
	},
	computed: {
		requestResponseHeadersFormatted: function() {
			const res = [];
			for (var k in this.requestResponseHeaders) {
				let v = this.requestResponseHeaders[k];
				res.push(`${k}: ${v}`);
			}

			return res.join('\r\n');
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

		uninitialize: function() {
			this.userManager.removeUser().then(() => {
				this.userNamanger = null;
				this.user = null;
				this.initialized = false;
			});
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

		removeUser: async function() {
			return this.userManager.removeUser().then(() => {
				console.log('xxx', arguments);
				this.user = null;
			});
		},

		gcGet: async function(url, options={}) {
			this.requestResponse = '';
			this.requestStatus = null;
			this.requestNextLink = null;
			const start = new Date();

			return this.$http.get(url, options).then(response => {
				// Whoohoo success.
				this.requestResponse = response.body;
				this.requestResponseHeaders = response.headers.map;
				this.requestStatus = {
					success: response.status >= 200 && response.status < 300,
					code: response.status,
					duration: (new Date()) - start
				};
				return response.body;
			}).catch(response => {
				this.requestResponse = response.body;
				this.requestResponseHeaders = response.headers.map;
				this.requestStatus = {
					success: false,
					code: response.status || 0,
					duration: (new Date()) - start
				};
				return response.body;
			});
		},

		gcPost: async function(url, body, options={}) {
			this.requestStatus = null;
			const start = new Date();

			return this.$http.post(url, body, options).then(response => {
				// Whooho success.
				this.requestStatus = {
					success: response.status >= 200 && response.status < 300,
					code: response.status,
					duration: (new Date()) - start
				};
				return response.body;
			}).catch(response => {
				this.requestStatus = {
					success: false,
					code: response.status,
					duration: (new Date()) - start
				};
				return response.body;
			});
		},

		changeRequestMode: function(mode) {
			this.requestTab = mode;
			this.requestEndpoint = '';
		},

		changeResponseMode: function(mode) {
			this.responseTab = mode;
		},

		runRequest: async function(discard) {
			const endpoint = this.requestEndpoint;
			if (!endpoint) {
				return;
			}

			let key = null;
			if (!discard) {
				key = endpoint;
				if (key.indexOf('/me/') === 0) {
					key = key.substr(4);
				}
				if (key.indexOf('/') === 0) {
					key = key.substr(1);
				}
			}

			return this.doRequest(this.apiPrefix + endpoint, key);
		},

		doRequest: async function(url, requestKey) {
			console.info('run request', url, requestKey);

			return this.gcGet(url).then(response => {
				if (requestKey) {
					this.requestKey = requestKey;
					this.requestResults[requestKey] = response;

					if (response['@odata.nextLink']) {
						this.requestNextLink = response['@odata.nextLink'];
					}
				}

				return response;
			});
		},

		runRequestNextLink: async function() {
			return this.doRequest(this.requestNextLink, this.requestKey);
		},

		createTestMessage: async function() {
			this.createStatus = null;

			const payload = {
				subject: 'The standard Lorem Ipsum passage, used since the 1500s',
				toRecipients: [
					{
						name: 'Lorus Ipsis',
						address: 'user1@localhost'
					}
				],
				body: {
					contentType: 'text',
					content: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.'
				}
			}

			return this.gcPost(this.apiPrefix + '/me/sendMail', {
				message: payload
			}).then(response => {
				this.createStatus = this.requestStatus;

				return response;
			});
		}
	}
});
