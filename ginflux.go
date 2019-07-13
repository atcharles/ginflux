package ginflux

const Version = "v0.0.1"

type (
	Engine struct {
		pool    *oPool
		session *Session
	}
)

func (e *Engine) DB(name string) *db {
	return e.Session().NewDB(name)
}

func (e *Engine) NewDB(name string) (dbInstance *db, err error) {
	sn, er := e.newSession()
	if er != nil {
		err = er
		return
	}
	dbInstance = sn.NewDB(name)
	return
}

func (e *Engine) Session() *Session {
	return &Session{engine: e}
}

func (e *Engine) NewSession() (session *Session, err error) {
	return e.newSession()
}

func (e *Engine) newSession() (session *Session, err error) {
	session = &Session{engine: e}
	session.client, err = e.Acquire()
	return
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
	session, er := e.newSession()
	if er != nil {
		err = er
		return
	}
	for _, value := range beans {
		if err = session.createDB(value.DBName); err != nil {
			return
		}
		if err = session.CreateRetentionPolicy(value); err != nil {
			return
		}
	}
	session.Release()
	return
}
