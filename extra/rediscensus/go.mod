module github.com/peilei-hub/redis/extra/rediscensus/v8

go 1.15

replace github.com/peilei-hub/redis => ../..

replace github.com/peilei-hub/redis/extra/rediscmd/v8 => ../rediscmd

require (
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/peilei-hub/redis v8.11.6-proxy+incompatible
	github.com/peilei-hub/redis/extra/rediscmd/v8 v8.11.6-proxy+incompatible
	go.opencensus.io v0.23.0
)
