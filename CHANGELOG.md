# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.4.4]
- remove extra rpm config files [#43](https://github.com/xmidt-org/themis/pull/43)
- add JWK support [#48](https://github.com/xmidt-org/themis/pull/48)
- add pprof support [#50](https://github.com/xmidt-org/themis/pull/50)
- add content negotation for /keys [#53](https://github.com/xmidt-org/themis/pull/53)

## [v0.4.3]
- fix rpm spec file for epel 8 [#42](https://github.com/xmidt-org/themis/pull/42)

## [v0.4.2]
- fix rpm spec file, fix changelog formatting [#41](https://github.com/xmidt-org/themis/pull/41)

## [v0.4.1]
- added docker automation
- updated release pipeline to use travis
- added specialized partner id logic [#40](https://github.com/xmidt-org/themis/pull/40)

## [v0.4.0]
- Removed the required option for claims and metadata obtained from HTTP requests

## [v0.3.2]
- Add versioning to themis binaries 

## [v0.3.1]
- Added a custom xhttpserver.Listener type
- Added MaxConcurrentRequests enforcement driven by configuration
- ConstantHandler for static HTTP transaction responses
- Busy decorator for enforcing MaxConcurrentRequests

## [v0.3.0]
- Allow metrics and health servers to be disabled
- Allow only a claims server to be configured
- Require an issuer server if a keys server is configured, and vice versa

## [v0.2.1]
- Use metrics namespace from config

## [v0.2.0]
- added configurable and application-injectable peer verification for MTLS

## [v0.1.1]
- Use new paths for systemd start

## [v0.1.0]
- Added logic to create RPMs per themis running mode

## [v0.0.3]
- updated Makefile
- updated conf directory
- Refactored config and xlog packages to remove some magic and makes things more obvious
- Allow named HTTP client components
- Simplify HTTP client/server component providers

## [v0.0.2]
- Fixed issues with building themis as a module

## [v0.0.1]
- Rename from xmidt-issuer to themis to follow the naming convention
- Dev mode
- Uber/fx style provders
- MTLS support
- Remote claims support
- Request logging
- Integrated server logging
- Full support for claims specified in requests
- Optional claims server that simply returns a JSON payload of the claims
- Time-based claims can be disabled
- Both the issue and claims servers can be disabled
- Integrated health via InvisionApp/go-health
- Converted to a go module: github.com/xmidt-org/themis


[Unreleased]: https://github.com/xmidt-org/themis/compare/v0.4.4...HEAD
[v0.4.4]: https://github.com/xmidt-org/themis/compare/v0.4.3...v0.4.4
[v0.4.3]: https://github.com/xmidt-org/themis/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/xmidt-org/themis/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/xmidt-org/themis/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/xmidt-org/themis/compare/v0.3.2...v0.4.0
[v0.3.2]: https://github.com/xmidt-org/themis/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/xmidt-org/themis/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/xmidt-org/themis/compare/v0.2.1...v0.3.0
[v0.2.1]: https://github.com/xmidt-org/themis/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/xmidt-org/themis/compare/v0.1.1...v0.2.0
[v0.1.1]: https://github.com/xmidt-org/themis/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/xmidt-org/themis/compare/v0.0.3...v0.1.0
[v0.0.3]: https://github.com/xmidt-org/themis/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/xmidt-org/themis/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/xmidt-org/themis/compare/v0.0.0...v0.0.1
