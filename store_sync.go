package godux

import (
	"sync"
)

type SyncStore struct {
	state          interface{}
	reducers       []ReducerHandler
	subscibers     []SubscribeHandler
	middlewares    []MiddlewareHandler
	locker         sync.Mutex
	haveMiddleware bool
}

func NewSyncStore(initState interface{}, reducers []ReducerHandler) Store {
	store := &SyncStore{
		state:          initState,
		reducers:       reducers,
		subscibers:     make([]SubscribeHandler, 0),
		middlewares:    make([]MiddlewareHandler, 0),
		haveMiddleware: false,
	}

	return store
}

func (s *SyncStore) GetState() interface{} {
	s.locker.Lock()
	defer s.locker.Unlock()
	return s.state
}

func (s *SyncStore) Dispatch(actionCreator ActionCreator) {
	actionCreator(s.dispatch, s.GetState)
}

func (s *SyncStore) Subscribe(handler SubscribeHandler) {
	s.subscibers = append(s.subscibers, handler)
}

func (s *SyncStore) ApplyMiddleware(handler MiddlewareHandler) {
	s.middlewares = append(s.middlewares, handler)
	s.haveMiddleware = true
}

func (s *SyncStore) Close() {}

func (s *SyncStore) dispatch(action Action) {
	if s.haveMiddleware {
		for _, handler := range s.middlewares {
			handler(s.dispatch, action, s.next)
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

func (s *SyncStore) next(action Action) {
	s.locker.Lock()
	for _, handler := range s.reducers {
		s.state = handler(s.state, action)
	}
	s.locker.Unlock()
}
