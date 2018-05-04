package godux

import "sync"

type SubscribeHandler func(state interface{}, action Action)
type ReducerHandler func(state interface{}, action Action) interface{}
type Dispatch func(action Action)
type MiddlewareHandler func(dispatch Dispatch, action Action, next Next)
type Next func(action Action)
type GetState func() interface{}

type subData struct {
	action Action
	state  interface{}
}

type Store interface {
	GetState() interface{}
	Subscribe(handler SubscribeHandler)
	Dispatch(ActionCreator)
	ApplyMiddleware(handler MiddlewareHandler)
	Close()
}

type DefaultStore struct {
	state          interface{}
	reducers       []ReducerHandler
	subscibers     []SubscribeHandler
	middlewares    []MiddlewareHandler
	dispatcher     chan Action
	chanSubData    chan subData
	closeDispatch  chan struct{}
	closeSub       chan struct{}
	locker         sync.Mutex
	haveMiddleware bool
}

func NewStore(initState interface{}, reducers []ReducerHandler) Store {
	store := &DefaultStore{
		state:          initState,
		reducers:       reducers,
		subscibers:     make([]SubscribeHandler, 0),
		dispatcher:     make(chan Action, 1000),
		chanSubData:    make(chan subData, 1000),
		middlewares:    make([]MiddlewareHandler, 0),
		closeDispatch:  make(chan struct{}, 1),
		closeSub:       make(chan struct{}, 1),
		haveMiddleware: false,
	}

	go store.watchDispatch()
	go store.watchSub()

	return store
}

func (s *DefaultStore) GetState() interface{} {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.state
}

func (s *DefaultStore) Dispatch(actionCreator ActionCreator) {
	actionCreator(s.dispatch, s.GetState)
}

func (s *DefaultStore) Subscribe(handler SubscribeHandler) {
	s.subscibers = append(s.subscibers, handler)
}

func (s *DefaultStore) ApplyMiddleware(handler MiddlewareHandler) {
	s.middlewares = append(s.middlewares, handler)
	s.haveMiddleware = true
}

func (s *DefaultStore) Close() {
	s.closeDispatch <- struct{}{}
	s.closeSub <- struct{}{}
}

func (s *DefaultStore) watchDispatch() {
	for {
		select {
		case <-s.closeDispatch:
			return
		case action := <-s.dispatcher:
			if s.haveMiddleware {
				for _, handler := range s.middlewares {
					handler(s.dispatch, action, s.next)
				}
			} else {
				s.next(action)
			}
		}
	}
}

func (s *DefaultStore) watchSub() {
	for {
		select {
		case <-s.closeSub:
			return
		case st := <-s.chanSubData:
			for _, handler := range s.subscibers {
				handler(st.state, st.action)
			}
		}
	}
}

func (s *DefaultStore) dispatch(action Action) {
	s.dispatcher <- action
}

func (s *DefaultStore) next(action Action) {
	s.locker.Lock()
	for _, handler := range s.reducers {
		s.state = handler(s.state, action)
	}
	s.locker.Unlock()
	s.chanSubData <- subData{action, s.state}
}
