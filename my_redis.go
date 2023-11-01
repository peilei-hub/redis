package redis

import "errors"

type RedisClient interface {
	UniversalClient
}

func NewRedisClient(opts *UniversalOptions) (RedisClient, error) {
	if opts == nil {
		return nil, errors.New("opts is nil")
	}

	var cli RedisClient
	var err error
	switch opts.DriverType {
	case StandAlone:
		cli = NewClient(opts.Simple())
	case Cluster:
		cli = NewClusterClient(opts.Cluster())
	case Sentinel:
		cli = NewFailoverClient(opts.Failover())
	case Proxy:
		cli, err = NewProxyClient(opts.Proxy())
	}
	if err != nil {
		return nil, err
	}

	return cli, nil
}
