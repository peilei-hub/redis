module github.com/peilei-hub/redis/extra/redisotel/v8

go 1.15

replace github.com/peilei-hub/redis => ../..

replace github.com/peilei-hub/redis/extra/rediscmd/v8 => ../rediscmd

require (
	github.com/peilei-hub/redis v8.11.6-proxy+incompatible
	github.com/peilei-hub/redis/extra/rediscmd/v8 v8.11.6-proxy+incompatible
	go.opentelemetry.io/otel v1.5.0
	go.opentelemetry.io/otel/sdk v1.4.1
	go.opentelemetry.io/otel/trace v1.5.0
)
