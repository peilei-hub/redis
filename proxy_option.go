package redis

import (
	"context"
	"time"
)

const (
	defaultAutoLoadInterval = 30 * time.Second
)

func DefaultGetProxies(ctx context.Context) []string {
	return []string{"localhost:6379"}
}

type ProxyOption struct {
	*Options
	AutoLoadProxy    bool
	AutoLoadInterval time.Duration
	GetProxies       GetProxies
}

func (opt *ProxyOption) init() {
	if opt == nil {
		*opt = ProxyOption{}
	}
	if opt.Options == nil {
		opt.Options = &Options{}
		opt.Options.init()
	}

	if opt.GetProxies == nil {
		opt.AutoLoadProxy = false
		opt.GetProxies = DefaultGetProxies
	}

	if opt.AutoLoadProxy {
		if opt.AutoLoadInterval < defaultAutoLoadInterval {
			opt.AutoLoadInterval = defaultAutoLoadInterval
		}
	}
}
