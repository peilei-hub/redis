package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8/internal/pool"
	"time"
)

type GetProxies func(ctx context.Context) []string

func DefaultGetProxies(ctx context.Context) []string {
	return []string{"localhost:6379"}
}

type ProxyClient struct {
	*Client
	proxyPool *ProxyPool
}

func NewProxyClient(getProxies GetProxies, opt *ProxyOption) (*ProxyClient, error) {
	opt.init()
	ctx := context.Background()
	if getProxies == nil {
		getProxies = DefaultGetProxies
		opt.AutoLoadProxy = false
	}

	cli := &ProxyClient{
		Client: NewClient(opt.Options),
	}

	cli.SetSelfDefineGetConn(cli.getConn)
	cli.SetSelfDefineReleaseConn(cli.releaseConn)

	ch := make(chan []string, 1)
	proxyList := getProxies(ctx)

	proxyPool, err := newProxyPool(opt, proxyList, ch)
	if err != nil {
		return nil, err
	}
	cli.proxyPool = proxyPool

	if opt.AutoLoadProxy {
		go autoLoadProxy(ctx, getProxies, ch, opt.AutoLoadInterval)
	}

	return cli, nil
}

func (p *ProxyClient) getConn(ctx context.Context) (*pool.Conn, error) {
	cn, err := p.proxyPool.GetConn(ctx)
	if err != nil {
		return nil, err
	}

	if cn.Inited {
		return cn, nil
	}

	if err := p.initConn(ctx, cn); err != nil {
		p.proxyPool.ReleaseConn(ctx, cn, err)
		if err := errors.Unwrap(err); err != nil {
			return nil, err
		}
		return nil, err
	}

	return cn, nil
}

func (p *ProxyClient) releaseConn(ctx context.Context, cn *pool.Conn, err error) {
	p.proxyPool.ReleaseConn(ctx, cn, err)
}

func autoLoadProxy(ctx context.Context, getProxies GetProxies, ch chan []string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ch <- getProxies(ctx)
	}
}
