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

		apiPrefix: '/api/gc/v1',
		pubsPrefix: '/api/pubs/v1',

		requestTab: '',
		requestKey: null,
		requestEndpoint: '',
		requestResponse: '',
		requestResponseJSON: undefined,
		requestResponseEditor: null,
		requestResponseHeaders: null,
		requestNextLink: null,
		bodyEditor: null,

		requestMode: {
			default: true
		},
		requestStatus: null,
		requestResults: {},
		requestResult: null,
		requestContext: 'me/calendar/events',

		responseTab: '',
		responseMode: {
			default: true
		},

		createStatus: null,
		subscriptionStatus: null,
		subscriptions: {},

		webhook: null,
		webhookClientState: 'whcs-' + new Date().getTime(),

		pubs: null,
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

		this.$nextTick(() => {
			// Code editor.
			this.requestResponseEditor = ace.edit(this.$refs.requestResponseEditor);
			this.requestResponseEditor.getSession().setMode("ace/mode/json");
			this.requestResponseEditor.setReadOnly(true);
			this.requestResponseEditor.setShowPrintMargin(false);
			this.requestResponseEditor.$blockScrolling = Infinity;
			// HTML Editor.
			this.bodyEditor = new Quill(this.$refs.bodyEditor, {
				debug: 'info',
				modules: {
					toolbar: true,
					clipboard: true,
				},
				theme: 'snow'
			});
			this.bodyEditor.disable();
		});

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
				Vue.http.headers.common['Authorization'] = user.token_type + ' ' + user.access_token;
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
		},
		requestResponse: function(val) {
			if (val.trim()) {
				this.requestResponseEditor.setValue(val);
			} else {
				this.requestResponseEditor.setValue('');
			}
			this.requestResponseEditor.clearSelection();
		},
		requestResponseJSON: function(val) {
			if (val && val.body !== undefined) {
				let content = val.body.content;
				console.debug(`raw body (${val.body.contentType})`, content);

				switch (val.body.contentType) {
					case 'text':
						const reader = new commonmark.Parser();
						const writer = new commonmark.HtmlRenderer({
							safe: true,
							smart: true,
							softbreak: '<br/>',
						});
						const parsed = reader.parse(content);
					 	content = writer.render(parsed);

						console.debug('converted body', content);
						break;

					case 'html':
						// Yeah .. :/
						break;
				}

				const clean = DOMPurify.sanitize(content, {
					FORBID_TAGS:    ['svg'],
					WHOLE_DOCUMENT: false,
				});

				console.debug('clean body', clean);

				this.bodyEditor.clipboard.dangerouslyPasteHTML(clean, 'api');
				this.bodyEditor.enable();
			} else {
				this.bodyEditor.disable();
				this.bodyEditor.clipboard.dangerouslyPasteHTML('', 'api');

			}
		},
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
				scope: 'openid profile email kopano/gc',
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
				this.updateUser(user);
			});
			mgr.events.addUserUnloaded(() => {
				console.log('user unloaded');
				this.updateUser(null);
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
						console.log('retrying silent renew of user', user);
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
				this.updateUser(null);
				this.initialized = false;
			});
		},

		startAuthentication: async function() {
			return this.userManager.signinRedirect();
		},

		completeAuthentication: async function() {
			return this.userManager.signinRedirectCallback().then(user => {
				return this.updateUser(user);
			});
		},

		getUser: async function() {
			return this.userManager.getUser().then(async (user) => {
				return this.updateUser(user);
			}).catch((err) => {
				console.error('failed to get user', err);
				return this.updateUser(null);
			});
		},

		removeUser: async function() {
			return this.userManager.removeUser().then(() => {
				return this.updateUser(null);
			});
		},

		updateUser: async function(user) {
			console.log('user updated', user);
			if (user) {
				Vue.http.headers.common['Authorization'] = user.token_type + ' ' + user.access_token;
				if (!this.webhook) {
					this.registerWebhook();
				}
				if (!this.pubs) {
					Pubs.init({authorizationValue: user.access_token, authorizationType: user.token_type});
					this.connectPubs();
				}
			}
			this.user = user;

			return user;
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
				this.requestResponseHeaders = response.headers.map;
				this.requestStatus = {
					success: response.status >= 200 && response.status < 300,
					code: response.status,
					duration: (new Date()) - start
				};

				this.requestResponse = response.bodyText;
				if (response.headers.get('content-type').indexOf('application/json') === 0) {
					return response.json().then(data => {
						this.requestResponseJSON = data;
						return data;
					});
				} else {
					this.requestResponseJSON = undefined;
					return response.text();
				}
			}).catch(response => {
				response.text().then(t => {
					this.requestResponse = t;
				});
				this.requestResponseJSON = undefined;
				this.requestResponseHeaders = response.headers.map;
				this.requestStatus = {
					success: false,
					code: response.status || 0,
					duration: (new Date()) - start
				};
				return {};
			});
		},

		gcPost: async function(url, body, options={}) {
			this.requestStatus = null;
			const start = new Date();

			return this.$http.post(url, body, options).then(response => {
				this.requestResponseHeaders = response.headers.map;
				this.requestStatus = {
					success: response.status >= 200 && response.status < 300,
					code: response.status,
					duration: (new Date()) - start
				};

				this.requestResponse = response.bodyText;
				if (response.headers.get('content-type').indexOf('application/json') === 0) {
					return response.json();
				} else {
					return response.text();
				}
			}).catch(response => {
				response.text().then(t => {
					this.requestResponse = t;
				});
				this.requestResponseHeaders = response.headers.map;
				this.requestStatus = {
					success: false,
					code: response.status || 0,
					msg: ''+response,
					duration: (new Date()) - start
				};
				return {};
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

			let url = endpoint;
			if (url === '/me/calendar/calendarView') {
				url += '?startDateTime=2017-02-26T09:13:19.540&endDateTime=2018-03-05T09:13:19.540';
			}

			return this.doRequest(this.apiPrefix + url, key);
		},

		doRequest: async function(url, requestKey) {
			console.info('run request', url, requestKey);

			return this.gcGet(url).then(response => {
				this.requestResult = response;
				this.requestContext = '';
				if (response && response.hasOwnProperty('@odata.context')) {
					this.requestContext = response['@odata.context'].substr(this.apiPrefix.length + 1);
				}

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
						emailAddress: {
							name: this.requestResults.me.userPrincipalName,
							address: this.requestResults.me.mail
						}
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
		},

		createTestEvent: async function() {
			this.createStatus = null;

			const start = moment();
			const payload = {
				subject: 'The standard Lorem Ipsum passage, used since the 1500s',
				start: {
					dateTime: start
				},
				end: {
					dateTime: moment(start).add(1, 'hour')
				}
			};

			return this.gcPost(this.apiPrefix + '/me/calendar/events', payload).then(response => {
				this.createStatus = this.requestStatus;

				return response;
			});
		},

		createRandomContact: async function() {
			this.createStatus = null;

			const r = Math.random().toString(36)
			const payload = {
				displayName: `Some contact ${r}`,
				emailAddresses: [
					{
						name: `Some contact ${r} first email`,
						address: `some-contact-${r}-first-email@kopano.local`
					}
				]
			};

			return this.gcPost(this.apiPrefix + '/me/contactFolders/contacts/contacts', payload).then(response => {
				this.createStatus = this.requestStatus;

				return response;
			})
		},

		registerWebhook: async function() {
			return this.$http.post(this.pubsPrefix + '/webhook').then(response => {
				if (response.headers.get('content-type').indexOf('application/json') !== 0) {
					// Require JSON response, everything else is an error.
					throw response;
				}

				// Whoohoo success.
				return response.json();
			}).catch(response => {
				console.error('failed to register webhook', response);
				return {};
			}).then(async webhook => {
				this.webhook = webhook;
				console.info('webhook registered', webhook);
				if (this.pubs) {
					// Subscribe our webhooks topic.
					const topic = this.webhook.topic;
					console.info('pubs subscribing webhook topic', topic);
					await this.pubs.sub([topic]);
					console.log('pubs subscribed', topic);
				}
			});
		},

		connectPubs: async function() {
			const pubs = new Pubs();
			await pubs.connect();
			this.pubs = pubs;
			if (this.webhook && this.webhook.topic) {
				// Subscribe to webhook topic.
				const topic = this.webhook.topic;
				console.info('pubs subscribing webhook topic to newly created pubs', topic);
				pubs.onstreamevent = (event) => {
					console.log('pubs stream event', event.data, event.info, event);
				}
				await this.pubs.sub([topic]);
				console.log('pubs subscribed', topic);
			}
		},

		createSubscription: async function() {
			console.log('create subscription', this.requestContext);

			const resource = this.requestContext;

			const changeType = "created,updated,deleted";
			const expirationDateTime = new Date();
			const payload = {
				"changeType": changeType,
				"resource": resource,
				"expirationDateTime": expirationDateTime,
				"clientState": this.webhookClientState,
				"notificationUrl": qualifyURL(this.webhook.pubUrl),
			}

			this.subscriptionStatus = null;
			const start = new Date();
			return this.$http.post(this.apiPrefix + '/subscriptions', payload).then(response => {
				if (response.headers.get('content-type').indexOf('application/json') !== 0) {
					// Require JSON response, everything else is an error.
					throw response;
				}

				this.subscriptionStatus = {
					success: response.status >= 200 && response.status < 300,
					code: response.status,
					duration: (new Date()) - start
				};
				return response.json();
			}).catch(response => {
				console.error('failed to subscribe', response);
				this.subscriptionStatus = {
					success: false,
					code: response.status || 0,
					msg: ''+response,
					duration: (new Date()) - start
				};
				return {};
			}).then(subscription => {
				this.subscriptions[resource] = subscription;
				console.info('subscribed', resource, subscription);
			});
		}

	}
});
