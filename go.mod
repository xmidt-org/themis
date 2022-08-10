module github.com/xmidt-org/themis

go 1.12

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/InVisionApp/go-logger v1.0.1
	github.com/go-kit/kit v0.10.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/justinas/alice v1.2.0
	github.com/lestrrat-go/jwx v0.9.2
	github.com/prometheus/client_golang v1.4.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.8.0
	github.com/xmidt-org/candlelight v0.0.5
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.19.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.19.0
	go.uber.org/fx v1.13.0
	go.uber.org/multierr v1.8.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
