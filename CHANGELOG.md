# CHANGELOG

## Unreleased

- Allow to start without any plugin
- Implement initialization order for plugins
- Bump linter to v1.21.0
- Update 3rd party modules
- Add minimal version unit test
- Use Go modules instead of Go dep


## v0.13.2 (2019-11-07)

- Build with Go 1.13.4


## v0.13.1 (2019-10-31)

- Build with Go 1.13.3
- Fix typo on README


## v0.13.0 (2019-10-01)

- Build with Go 1.13
- Add basic metrics
- Ensure BASE folder in fmt and check targets
- Improve README


## v0.12.2 (2019-09-30)

- Rebuild with Go 1.12.10


## v0.12.1 (2019-08-26)

- Update Go Dep to 0.5.4
- Compile on stretch to ensure compatibility


## v0.12.0 (2019-08-22)

- Use sticky backend selection for grapi
- Add GRAPI link to top level README
- Add more GRAPI information to README
- Unify env variable names of test scripts
- Update README
- Add instructions to example scripts
- Update plugin docs with additional details where those were missing


## v0.11.0 (2019-07-26)

- Add healthcheck sub command


## v0.10.0 (2019-05-17)

- Use correct Dep tag
- Use updated golint
- Build with Go 1.12 and Dep 0.5.2
- Update libkcoidc to 0.6.0 and use its new flexible logger
- Improve access token python example
- Fixup python syntax
- Add OAuth2 python3 example
- Add OIDC Python Kopano console helper
- Improve wording and fix typos
- oidc-client-example.html edited online with Bitbucket
- Show details and print OK for make check


## v0.9.0 (2019-01-28)

- Ensure third party caddy dependency is a compatible revision
- Require kopano/pubs scope to access pubs API
- Require kopano/kvs scope to access kvs
- Log what URL got access denied
- Fixup invalid debug log formatting


## v0.8.2 (2019-01-24)

- Show Go version in Jenkins


## v0.8.1 (2019-01-24)



## v0.8.0 (2019-01-22)

- Bump base copyright years to 2019
- Update dep docs and use pinned version
- Add CORS support to pubs (disabled by default)
- Add CORS support for kvs (disabled by default)
- Allow to set server to test CORS / use different host
- Update Makefile to no longer require local pubsjs
- scripts: remove duplicate ProtectSystem line
- Use latest released pubs version
- Update to updates Pubs module API


## v0.7.0 (2018-11-20)

- Update Caddy and its dependencies to same as used in kweb
- Add kvs batch mode for create or update
- Add code coverage reporting to Jenkins
- Add unit tests for KV
- Ensure array response format when recuse is requested
- Add docs for KVS
- Ensure to recurse when requested even if only one result
- Limit number of database connections to 10
- Add support for DELETE
- Add minimal DB open unit tests
- Add configuration for kvs
- Include db migrations in dist tarball
- Implement flexible persistent key value store
- Fixup deprecation warning in Jenkins
- Update license ranger to version with dep support
- Replace Glide with Go deps
- Build with Go 1.11
- Update build checks


## v0.6.4 (2018-09-26)

- Remove obsolete use of external environment files


## v0.6.3 (2018-09-21)

- Update libkcoidc to 0.4.3


## v0.6.2 (2018-09-06)

- Fixup cast to pointer


## v0.6.1 (2018-09-06)

- Update libkcoidc to 0.4.1


## v0.6.0 (2018-09-06)

- Use scope check from kcoidc
- Update libkcoidc to 0.4.0
- Refactor header injection so plugins can do it
- grapi: Rename groupware-core plugin to grapi
- Increase no-file limit to infinite


## v0.5.1 (2018-08-21)

- Add setup subcommand to binscript


## v0.5.0 (2018-08-17)

- Run Jenkins with Go 1.10
- Add log-level to config and avoid double timestamp for systemd
- Split of serve subcommand into extra file
- Add commandline args for log output control
- groupware-core: Include socket folder check in retry
- Add systemd unit with runner script and config
- groupware-core: Add README
- Error out when plugins-path is not found
- Subscripe to deleted in example
- examples: define a working example


## v0.4.2 (2018-08-10)

- Add purify hooks and debug
- Update Quill to version 2.0.0-dev.2 for table support
- Add HTML editor to example (Quill)


## v0.4.1 (2018-07-31)

- Add support for kcoidc debug log


## v0.4.0 (2018-07-18)

- Disable obsolete groupware-core v0 API endpoints
- Add scope validation to groupware-core and pubs


## v0.3.0 (2018-07-16)

- Update Kopano GC REST API support to v1
- Update pubjs in playground example


## v0.2.0 (2018-04-13)

- Implement injection of username header
- Add Makefile to example app
- Ignore test files
- Add helpers for testing
- Add docs for -iss parameter


## v0.1.0 (2018-03-19)

- Add OIDC discovery to validat access tokens
- Rename project to kapi for consistency
- pubs: Require minimal key size
- Fix format logging
- Add wrapper for hey for easy testing
- Only fetch realistic fields
- Add calendarView request to example webapp
- Add calendarView scenario
- Use pubjs in example
- Add 3rd party license information


## v0.0.1 (2018-02-14)

- Add Jenkinsfile
- example: Add pubs support to example webapp
- pubs: Add authentication for streams and logging
- Add auth submodule to support context base auth
- Update socket glob for KC changes
- Fixup: update README
- Add pubs plugin for minimal pubsub and webhooks
- Disable request HTTP log by default
- Add molotov test scenario
- gc: Implement subscriptions delegation
- gc: Only use mfr sockets for REST backend
- gc: Add support for CORS (disabled by default)
- Improve error behavior when no JSON is received
- Update sendMail payload to use emailAddress
- Do not run linter on make
- Use ACE editor for JSON display
- /me/contacts POST is broken
- Always show scrollbars
- Add contact and event creation
- Add configuration for OIDC clientID
- Fixup README
- Implement OIDC silent renew
- Update example client to new gc API endpoints
- No longer rewrite jc API URL
- Use Base64 encoded UserEntryID
- Add Caddyfile for development
- Ignore dist and Caddyfile
- Avoid stuttering kopano-kopano.., package name is now kopano-api
- Move httpproxy implementation to submodule
- Add meta data to server and plugins
- Move groupware core endpoint to plugin
- Implement plugin interface
- mplement simple oidc web app with oidc-client-js
- Allow 8 parallel connection to proxy workers
- Implement parsing of Bearer access tokens
- Fixup: typos in README
- Fixup: strip path for proxy requests better
- Fixup: strip path for proxy requests
- Initial commit

