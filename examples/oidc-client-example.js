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
		silentRenew: true,
		iss: '',
		clientID: '',
		prompt: null,

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
		console.info('welcome to Kopanpo API oidc-client example');

		const queryValues = parseParams(location.search.substr(1));
		console.log('URL query values on load', queryValues);

		let iss = queryValues.iss;
		if (!iss) {
			iss = localStorage.getItem('oidc-client-example.iss');
		}
		let clientID = queryValues.client_id;
		if (!clientID) {
			clientID = localStorage.getItem('oidc-client-example.client_id');
		}
		if (!clientID) {
			clientID = "oidc-client-example.js";
		}

		this.iss = iss;
		this.clientID = clientID;
		this.prompt = '';

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
					return this.getUser();
				}).then(user => {
					console.info('initialized', user, this.isLoggedIn);
					if (user && !this.isLoggedIn) {
						return this.userManager.signinSilent();
					}
					return Promise.resolve(user);
				}).then(user => {
					console.log('initialized phase 2', user, this.isLoggedIn);
					if (user && !this.isLoggedIn) {
						return this.removeUser();
					}
					return Promise.resolve(user);
				});
			});
		}
	},
	watch: {
		iss: function(val) {
			if (val) {
				localStorage.setItem('oidc-client-example.iss', val);
			} else {
				localStorage.removeItem('oidc-client-example.iss');
			}
		},
		clientID: function(val) {
			if (val) {
				localStorage.setItem('oidc-client-example.client_id', val);
			} else {
				localStorage.removeItem('oidc-client-example.client_id');
			}
		},
		user: function(user) {
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
		},

		isLoggedIn: function() {
			return this.user != null && !this.user.expired;
		}
	},
	methods: {
		initialize: function() {
			const callbackURI = window.location.href.split('#')[0] + '#callback';

			const mgr = new Oidc.UserManager({
				authority: this.iss,
				client_id: this.clientID,
				redirect_uri: callbackURI,
				response_type: 'id_token token',
				scope: 'openid profile email',
				loadUserInfo: true,
				silent_redirect_uri: qualifyURL('./oidc-silent-refresh.html'),
				accessTokenExpiringNotificationTime: 120,
				automaticSilentRenew: this.silentRenew,
				includeIdTokenInSilentRenew: true,
				prompt: this.prompt
			});
			mgr.events.addAccessTokenExpiring(() => {
				console.log('token expiring');
			});
			mgr.events.addAccessTokenExpired(() => {
				console.log('access token expired');
				mgr.removeUser();
			});
			mgr.events.addUserLoaded(user => {
				console.log('user loaded', user);
				this.user = user;
			});
			mgr.events.addUserUnloaded(() => {
				console.log('user unloaded');
				this.user = null;
			});
			mgr.events.addSilentRenewError(err => {
				console.log('user silent renew error', err.error);
				if (err && err.error === 'interaction_required') {
					// Handle the hopeless.
					return;
				}

				setTimeout(() => {
					if (!this.silentRenew) {
						return;
					}
					console.log('retrying silent renew');
					mgr.getUser().then(user => {
						console.log('retrying silent renew of user', user, user.expired);
						if (user && !user.expired) {
							mgr.startSilentRenew();
						} else {
							console.log('remove user as silent renew has failed to renew in time');
							mgr.removeUser();
						}
					});
				}, 5000);
			});
			mgr.events.addUserSignedOut(() => {
				console.log('user signed out');
			});

			this.userManager = mgr;

			if (window.location.href.indexOf(callbackURI) === 0) {
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
				this.user = null;
			});
		},

		querySessionStatus: async function() {
			return this.userManager.querySessionStatus().then(sessionStatus => {
				console.log('sessionStatus', sessionStatus);
				return sessionStatus;
			});
		},

		startSilentRenew: function() {
			this.silentRenew = true;
			this.userManager.startSilentRenew();
		},

		stopSilentRenew: function() {
			this.silentRenew = false;
			this.userManager.stopSilentRenew();
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
						name: this.requestResults.me.userPrincipalName,
						address: this.requestResults.me.mail
					}
				],
				body: {
					contentType: 'text',
					content: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.'
				}
			};

			return this.gcPost(this.apiPrefix + '/me/sendMail', {
				message: payload
			}).then(response => {
				this.createStatus = this.requestStatus;

				return response;
			});
		}
	}
});
