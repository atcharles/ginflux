package ginflux

//Version ...
const Version = "v0.0.1"

type (
	//Engine ...
	Engine struct {
		pool    *OPool
		session *Session
	}
)

//DB ...
func (e *Engine) DB(name string) *Database {
	return e.Session().NewDB(name)
}

//NewDB ...
func (e *Engine) NewDB(name string) (dbInstance *Database, err error) {
	sn, er := e.newSession()
	if er != nil {
		err = er
		return
	}
	dbInstance = sn.NewDB(name)
	return
}

//Session ...
func (e *Engine) Session() *Session {
	return &Session{engine: e}
}

//NewSession ...
func (e *Engine) NewSession() (session *Session, err error) {
	return e.newSession()
}

func (e *Engine) newSession() (session *Session, err error) {
	session = &Session{engine: e}
	session.client, err = e.Acquire()
	return
}

//NewEngine ...
func NewEngine(opts Options) (eg *Engine, err error) {
	pool := NewOPool(opts)
	var oc *OClient
	oc, err = pool.Acquire()
	if err != nil {
		return
	}
	defer oc.Release()
	eg = &Engine{pool: pool, session: &Session{}}
	eg.session.engine = eg
	return
}

//Acquire ...
func (e *Engine) Acquire() (cl *OClient, err error) {
	return e.pool.Acquire()
}

//SyncDB ...
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
