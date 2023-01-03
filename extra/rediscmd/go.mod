module github.com/peilei-hub/redis/extra/rediscmd/v8

go 1.15

replace github.com/peilei-hub/redis => ../..

require (
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/peilei-hub/redis v8.11.6-proxy+incompatible
)
