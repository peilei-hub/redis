package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8/internal/pool"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type ProxyPool struct {
	opts     *ProxyOption
	poolList []string
	poolMap  map[string]*pool.ConnPool
	ch       <-chan []string
	index    uint32
	sync.RWMutex
}

func newProxyPool(proxyOptions *ProxyOption, proxyList []string, ch <-chan []string) (*ProxyPool, error) {
	proxies := removeDuplicateProxy(proxyList)
	if len(proxies) == 0 {
		return nil, errors.New("proxyList len is 0")
	}

	proxyPool := &ProxyPool{
		opts: proxyOptions,
		ch:   ch,
	}

	poolMap := make(map[string]*pool.ConnPool)
	for _, proxyAddr := range proxies {
		opts := proxyOptions.Options.clone()
		opts.Addr = proxyAddr

		poolMap[proxyAddr] = newConnPool(opts)
	}
	proxyPool.poolMap = poolMap
	proxyPool.poolList = getKeys(poolMap)

	if proxyOptions.AutoLoadProxy {
		go proxyPool.autoLoadProxy()
	}

	return proxyPool, nil
}

func getKeys(poolMap map[string]*pool.ConnPool) []string {
	result := make([]string, 0)
	for k := range poolMap {
		result = append(result, k)
	}
	return result
}

func (p *ProxyPool) getConnPool() (*pool.ConnPool, bool) {
	p.Lock()
	defer p.Unlock()
	index := atomic.AddUint32(&p.index, 1) % uint32(len(p.poolList))

	connPool, ok := p.poolMap[p.poolList[index]]
	return connPool, ok
}

func (p *ProxyPool) GetConn(ctx context.Context) (*pool.Conn, error) {
	connPool, ok := p.getConnPool()
	if !ok {
		log.Fatal("get nil connPool")
	}

	return connPool.Get(ctx)
}

func (p *ProxyPool) ReleaseConn(ctx context.Context, cn *pool.Conn, err error) {
	addr := cn.RemoteAddr().String()
	p.RLock()
	connPool, ok := p.poolMap[addr]
	p.RUnlock()
	if !ok {
		log.Fatal("error")
		return
	}
	if isBadConn(err, false, addr) {
		connPool.Remove(ctx, cn, err)
	} else {
		connPool.Put(ctx, cn)
	}
}

func (p *ProxyPool) autoLoadProxy() {
	for {
		select {
		case proxies := <-p.ch:
			p.updateProxies(proxies)
		}
	}
}

func (p *ProxyPool) updateProxies(proxyList []string) {
	if len(proxyList) == 0 {
		log.Fatal("proxies length is 0")
		return
	}

	proxies := removeDuplicateProxy(proxyList)

	if !p.proxyChanged(proxies) {
		return
	}

	newPoolMap := make(map[string]*pool.ConnPool)
	for _, proxyAddr := range proxies {
		connPool, ok := p.poolMap[proxyAddr]
		if ok {
			newPoolMap[proxyAddr] = connPool
			continue
		}

		option := p.opts.Options.clone()
		option.Addr = proxyAddr

		connPool = newConnPool(p.opts.Options)
		newPoolMap[proxyAddr] = connPool
	}

	for key := range p.poolMap {
		connPool, ok := newPoolMap[key]
		if !ok {
			go laterClose(key, connPool)
		}
	}

	poolList := getKeys(newPoolMap)

	p.Lock()
	p.poolMap = newPoolMap
	p.poolList = poolList
	p.Unlock()
}

func (p *ProxyPool) proxyChanged(proxies []string) bool {
	if len(p.poolMap) != len(proxies) {
		return true
	}

	for _, proxyAddr := range proxies {
		if _, ok := p.poolMap[proxyAddr]; !ok {
			return true
		}
	}

	return false
}

func laterClose(addr string, connPool *pool.ConnPool) {
	time.Sleep(5 * time.Second)
	err := connPool.Close()
	if err != nil {
		log.Fatalf("close conn error, addr: %s, err: %v", addr, err)
	}
}

func removeDuplicateProxy(proxyList []string) []string {
	strMap := make(map[string]struct{})
	for _, s := range proxyList {
		strMap[s] = struct{}{}
	}

	result := make([]string, 0, len(strMap))
	for k := range strMap {
		result = append(result, k)
	}
	return result
}
