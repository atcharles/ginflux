package ginflux

import (
	"errors"
	"sync"
	"time"

	ic "github.com/influxdata/influxdb1-client/v2"
)

//NewOPool ...
func NewOPool(opt Options) *OPool {
	o := &OPool{
		lock:        new(sync.RWMutex),
		opt:         opt,
		currentOpen: 0,
	}
	if o.opt.MinOpen > o.opt.MaxOpen {
		panic("minOpen 不能大于 maxOpen")
	}
	if o.opt.MinOpen == 0 {
		o.opt.MinOpen = 10
	}
	if o.opt.MaxOpen == 0 {
		o.opt.MaxOpen = 100
	}
	if o.opt.GetTimeout == 0 {
		o.opt.GetTimeout = time.Millisecond * 100
	}
	o.pool = make(chan *OClient, o.opt.MaxOpen)
	return o
}

type (
	//Options ...
	Options struct {
		HttpConf ic.HTTPConfig
		//从池中取资源的超时时间
		GetTimeout time.Duration
		MinOpen    int
		MaxOpen    int
	}
	//OPool ...
	OPool struct {
		lock        *sync.RWMutex
		opt         Options
		currentOpen int64
		pool        chan *OClient
	}
	//OClient ...
	OClient struct {
		ic.Client
		lock *sync.RWMutex
		//是否正在使用
		inUsing bool
		alive   bool
		created time.Time
		op      *OPool
	}
)

//CurrentOpen ...
func (o *OPool) CurrentOpen() int64 {
	o.lock.RLock()
	n := o.currentOpen
	o.lock.RUnlock()
	return n
}

//Increment ...
func (o *OPool) Increment(n int64) {
	o.lock.Lock()
	o.currentOpen += n
	o.lock.Unlock()
}

var (
	//ErrMaxOpen ...
	ErrMaxOpen = errors.New("连接数超出最大限制")
	ErrTimeout = errors.New("从池中获取连接超时")
)

//Acquire ...
func (o *OPool) Acquire() (cl *OClient, err error) {
	return o.getOrCreateOne()
}

func (o *OPool) ping(oc *OClient) (err error) {
	oc.lock.Lock()
	defer oc.lock.Unlock()
	_, _, err = oc.Ping(time.Second * 1)
	if err == nil {
		oc.inUsing = true
		return
	}
	_ = oc.Close()
	o.Increment(-1)
	oc.alive = false
	return
}

func (o *OPool) newClient() (err error) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.currentOpen >= int64(o.opt.MaxOpen) {
		err = ErrMaxOpen
		return
	}
	oc := &OClient{
		lock:    new(sync.RWMutex),
		inUsing: false,
		alive:   true,
		created: time.Now(),
		op:      o,
	}
	oc.Client, err = ic.NewHTTPClient(o.opt.HttpConf)
	if err != nil {
		return
	}
	select {
	case o.pool <- oc:
		o.currentOpen++
	default:
	}
	return
}

func (o *OPool) getOrCreateOne() (oc *OClient, err error) {
	select {
	case oc = <-o.pool:
		err = o.ping(oc)
		return
	default:
		if err = o.newClient(); err != nil {
			return
		}
	}

	select {
	case oc = <-o.pool:
	case <-time.After(o.opt.GetTimeout):
		err = ErrTimeout
		return
	}
	return
}

//GetInfluxClient ...
func (o *OClient) GetInfluxClient() ic.Client {
	return o.Client
}

//Release ...
func (o *OClient) Release() {
	o.lock.Lock()
	defer o.lock.Unlock()
	if !o.alive {
		return
	}
	if !o.inUsing {
		return
	}
	o.inUsing = false

	select {
	case o.op.pool <- o:
	default:
	}
}
