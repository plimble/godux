package godux

import "sync"

type SubscribeHandler func(action Action)
type ReducerHandler func(state interface{}, action Action) interface{}
type Dispatch func(action Action)
type MiddlewareHandler func(dispatch Dispatch, action Action, next Next)
type Next func(action Action)
type GetState func() interface{}

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
	dispatcher     dispatcher
	close          chan struct{}
	locker         sync.Mutex
	haveMiddleware bool
}

func NewStore(initState interface{}, reducers []ReducerHandler) Store {
	store := &DefaultStore{
		state:          initState,
		reducers:       reducers,
		subscibers:     make([]SubscribeHandler, 0),
		dispatcher:     NewDispatcher(),
		middlewares:    make([]MiddlewareHandler, 0),
		close:          make(chan struct{}, 1),
		haveMiddleware: false,
	}

	go store.watch()

	return store
}

func (s *DefaultStore) GetState() interface{} {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.state
}

func (s *DefaultStore) Dispatch(actionCreator ActionCreator) {
	actionCreator(s.dispatcher.Dispatch, s.GetState)
}

func (s *DefaultStore) Subscribe(handler SubscribeHandler) {
	s.subscibers = append(s.subscibers, handler)
}

func (s *DefaultStore) ApplyMiddleware(handler MiddlewareHandler) {
	s.middlewares = append(s.middlewares, handler)
	s.haveMiddleware = true
}

func (s *DefaultStore) Close() {
	s.close <- struct{}{}
}

func (s *DefaultStore) watch() {
	for {
		select {
		case <-s.close:
			return
		case action := <-s.dispatcher.events:
			if s.haveMiddleware {
				for _, handler := range s.middlewares {
					handler(s.dispatcher.Dispatch, action, s.next)
					for _, handler := range s.subscibers {
						handler(action)
					}
				}
			} else {
				s.next(action)
				for _, handler := range s.subscibers {
					handler(action)
				}
			}
		}
	}
}

func (s *DefaultStore) next(action Action) {
	s.locker.Lock()
	for _, handler := range s.reducers {
		s.state = handler(s.state, action)
	}
	s.locker.Unlock()
}
