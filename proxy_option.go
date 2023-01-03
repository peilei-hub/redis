package redis

import (
	"time"
)

const (
	defaultAutoLoadInterval = 30 * time.Second
)

type ProxyOption struct {
	*Options
	AutoLoadProxy    bool
	AutoLoadInterval time.Duration
}

func (opt *ProxyOption) init() {
	if opt == nil {
		*opt = ProxyOption{}
	}
	if opt.Options == nil {
		opt.Options = &Options{}
		opt.Options.init()
	}

	if opt.AutoLoadProxy {
		if opt.AutoLoadInterval < defaultAutoLoadInterval {
			opt.AutoLoadInterval = defaultAutoLoadInterval
		}
	}
}
