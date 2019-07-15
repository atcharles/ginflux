package ginflux

import (
	"errors"
	"sync"
	"time"

	ic "github.com/influxdata/influxdb1-client/v2"
)

func NewOPool(opt Options) *oPool {
	o := &oPool{
		lock:        new(sync.RWMutex),
		opt:         opt,
		currentOpen: 0,
	}
	if o.opt.minOpen > o.opt.maxOpen {
		panic("minOpen 不能大于 maxOpen")
	}
	if o.opt.minOpen == 0 {
		o.opt.minOpen = 10
	}
	if o.opt.maxOpen == 0 {
		o.opt.maxOpen = 100
	}
	o.pool = make(chan *oClient, opt.maxOpen)
	if o.opt.getTimeout == 0 {
		o.opt.getTimeout = time.Second * 2
	}
	return o
}

type (
	Options struct {
		httpConf ic.HTTPConfig
		//从池中取资源的超时时间
		getTimeout time.Duration
		minOpen    int
		maxOpen    int
	}

	oPool struct {
		lock        *sync.RWMutex
		opt         Options
		currentOpen int64
		pool        chan *oClient
	}

	oClient struct {
		ic.Client
		lock *sync.RWMutex
		//是否正在使用
		inUsing bool
		alive   bool
		created time.Time
		op      *oPool
	}
)

func (o *oPool) CurrentOpen() int64 {
	o.lock.RLock()
	n := o.currentOpen
	o.lock.RUnlock()
	return n
}

func (o *oPool) Increment(n int64) {
	o.lock.Lock()
	o.currentOpen += n
	o.lock.Unlock()
}

var (
	ErrMaxOpen = errors.New("连接数超出最大限制")
	ErrInUsing = errors.New("连接正在使用,不能共享")
)

func (o *oPool) Acquire() (cl *oClient, err error) {
	now := time.Now()
	for {
		cl, err = o.getOrCreateOne()
		if err == nil {
			break
		}
		if time.Now().After(now.Add(o.opt.getTimeout)) {
			return
		}
	}
	return
}

func (o *oPool) getOrCreateOne() (oc *oClient, err error) {
	select {
	case oc = <-o.pool:
		oc.lock.RLock()
		if oc.inUsing && o.CurrentOpen() < int64(o.opt.maxOpen) {
			oc.lock.RUnlock()
			err = ErrInUsing
			return
		}
		oc.lock.RUnlock()
	default:
		if oc, err = o.newOClient(); err != nil {
			return
		}
	}
	_, _, err = oc.Ping(time.Second * 1)
	if err != nil {
		oc.lock.Lock()
		oc.alive = false
		oc.lock.Unlock()
		return
	}
	return
}

func (o *oPool) newOClient() (oc *oClient, err error) {
	oc = &oClient{
		lock:    new(sync.RWMutex),
		inUsing: true,
		alive:   true,
		created: time.Now(),
		op:      o,
	}
	oc.Client, err = ic.NewHTTPClient(o.opt.httpConf)
	select {
	case o.pool <- oc:
		o.Increment(1)
	default:
		err = ErrMaxOpen
		return
	}
	return
}

func (o *oClient) GetInfluxClient() ic.Client {
	return o.Client
}

func (o *oClient) Release() {
	if o == nil {
		return
	}
	o.lock.Lock()
	defer func() {
		o.lock.Unlock()
		o.op.pool <- o
	}()
	if !o.alive {
		return
	}
	if !o.inUsing {
		return
	}
	o.inUsing = false
}
