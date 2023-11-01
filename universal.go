package redis

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

type GetProxies func(ctx context.Context) []string

type DriverType string

const (
	// redis-server部署模式
	StandAlone DriverType = "stand_alone" // 单机版
	Sentinel   DriverType = "sentinel"    // 哨兵
	Cluster    DriverType = "cluster"     // 集群
	Proxy      DriverType = "proxy"       // redis代理
)

// UniversalOptions information is required by UniversalClient to establish
// connections.
type UniversalOptions struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addrs []string

	DriverType DriverType

	// for proxy
	GetProxies GetProxies

	AutoLoadProxy    bool
	AutoLoadInterval time.Duration

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int

	// Common options.

	Dialer    func(ctx context.Context, network, addr string) (net.Conn, error)
	OnConnect func(ctx context.Context, cn *Conn) error

	Username         string
	Password         string
	SentinelUsername string
	SentinelPassword string

	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// PoolFIFO uses FIFO mode for each node connection pool GET/PUT (default LIFO).
	PoolFIFO bool

	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration

	TLSConfig *tls.Config

	// Only cluster clients.

	MaxRedirects   int
	ReadOnly       bool
	RouteByLatency bool
	RouteRandomly  bool

	// The sentinel master name.
	// Only failover clients.

	MasterName string
}

// Cluster returns cluster options created from the universal options.
func (o *UniversalOptions) Cluster() *ClusterOptions {
	if len(o.Addrs) == 0 {
		o.Addrs = []string{"127.0.0.1:6379"}
	}

	return &ClusterOptions{
		Addrs:     o.Addrs,
		Dialer:    o.Dialer,
		OnConnect: o.OnConnect,

		Username: o.Username,
		Password: o.Password,

		MaxRedirects:   o.MaxRedirects,
		ReadOnly:       o.ReadOnly,
		RouteByLatency: o.RouteByLatency,
		RouteRandomly:  o.RouteRandomly,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:        o.DialTimeout,
		ReadTimeout:        o.ReadTimeout,
		WriteTimeout:       o.WriteTimeout,
		PoolFIFO:           o.PoolFIFO,
		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

// Failover returns failover options created from the universal options.
func (o *UniversalOptions) Failover() *FailoverOptions {
	if len(o.Addrs) == 0 {
		o.Addrs = []string{"127.0.0.1:26379"}
	}

	return &FailoverOptions{
		SentinelAddrs: o.Addrs,
		MasterName:    o.MasterName,

		Dialer:    o.Dialer,
		OnConnect: o.OnConnect,

		DB:               o.DB,
		Username:         o.Username,
		Password:         o.Password,
		SentinelUsername: o.SentinelUsername,
		SentinelPassword: o.SentinelPassword,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,

		PoolFIFO:           o.PoolFIFO,
		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

// Simple returns basic options created from the universal options.
func (o *UniversalOptions) Simple() *Options {
	addr := "127.0.0.1:6379"
	if len(o.Addrs) > 0 {
		addr = o.Addrs[0]
	}

	return &Options{
		Addr:      addr,
		Dialer:    o.Dialer,
		OnConnect: o.OnConnect,

		DB:       o.DB,
		Username: o.Username,
		Password: o.Password,

		MaxRetries:      o.MaxRetries,
		MinRetryBackoff: o.MinRetryBackoff,
		MaxRetryBackoff: o.MaxRetryBackoff,

		DialTimeout:  o.DialTimeout,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,

		PoolFIFO:           o.PoolFIFO,
		PoolSize:           o.PoolSize,
		MinIdleConns:       o.MinIdleConns,
		MaxConnAge:         o.MaxConnAge,
		PoolTimeout:        o.PoolTimeout,
		IdleTimeout:        o.IdleTimeout,
		IdleCheckFrequency: o.IdleCheckFrequency,

		TLSConfig: o.TLSConfig,
	}
}

func (o *UniversalOptions) Proxy() *ProxyOption {
	opts := &ProxyOption{
		Options:          o.Simple(),
		AutoLoadProxy:    o.AutoLoadProxy,
		AutoLoadInterval: o.AutoLoadInterval,
		GetProxies:       o.GetProxies,
	}

	opts.init()
	return opts
}

// --------------------------------------------------------------------

// UniversalClient is an abstract client which - based on the provided options -
// represents either a ClusterClient, a FailoverClient, or a single-node Client.
// This can be useful for testing cluster-specific applications locally or having different
// clients in different environments.
type UniversalClient interface {
	Cmdable
	Context() context.Context
	AddHook(Hook)
	Watch(ctx context.Context, fn func(*Tx) error, keys ...string) error
	Do(ctx context.Context, args ...interface{}) *Cmd
	Process(ctx context.Context, cmd Cmder) error
	Subscribe(ctx context.Context, channels ...string) *PubSub
	PSubscribe(ctx context.Context, channels ...string) *PubSub
	Close() error
	PoolStats() *PoolStats
}

var (
	_ UniversalClient = (*Client)(nil)
	_ UniversalClient = (*ClusterClient)(nil)
	_ UniversalClient = (*Ring)(nil)
)

// NewUniversalClient returns a new multi client. The type of the returned client depends
// on the following conditions:
//
// 1. If the MasterName option is specified, a sentinel-backed FailoverClient is returned.
// 2. if the number of Addrs is two or more, a ClusterClient is returned.
// 3. Otherwise, a single-node Client is returned.
func NewUniversalClient(opts *UniversalOptions) UniversalClient {
	if opts.MasterName != "" {
		return NewFailoverClient(opts.Failover())
	} else if len(opts.Addrs) > 1 {
		return NewClusterClient(opts.Cluster())
	}
	return NewClient(opts.Simple())
}
