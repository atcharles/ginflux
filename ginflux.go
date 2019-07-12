package ginflux

const Version = "v0.0.1"

type (
	Engine struct {
		pool    *oPool
		session *Session
	}
)

func (e *Engine) NewDB(name string) (dbInstance *db, err error) {
	return e.newSession().NewDB(name)
}

func (e *Engine) NewSession() *Session {
	return e.newSession()
}

func (e *Engine) newSession() *Session {
	return &Session{engine: e}
}

func NewEngine(opts Options) (eg *Engine, err error) {
	pool := NewOPool(opts)
	var oc *oClient
	oc, err = pool.Acquire()
	if err != nil {
		return
	}
	defer oc.Release()
	eg = &Engine{pool: pool, session: &Session{}}
	eg.session.engine = eg
	return
}

func (e *Engine) Acquire() (cl *oClient, err error) {
	return e.pool.Acquire()
}

func (e *Engine) SyncDB(beans ...RetentionPolicy) (err error) {
	for _, value := range beans {
		if err = e.newSession().CreateRetentionPolicy(value); err != nil {
			return
		}
	}
	return
}
