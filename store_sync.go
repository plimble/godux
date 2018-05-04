package godux

import (
	"sync"
)

type SyncStore struct {
	state          interface{}
	reducers       []ReducerHandler
	subscibers     []SubscribeHandler
	middlewares    []MiddlewareHandler
	chanSubData    chan subData
	closeSub       chan struct{}
	locker         sync.Mutex
	haveMiddleware bool
}

func NewSyncStore(initState interface{}, reducers []ReducerHandler) Store {
	store := &SyncStore{
		state:          initState,
		reducers:       reducers,
		subscibers:     make([]SubscribeHandler, 0),
		middlewares:    make([]MiddlewareHandler, 0),
		chanSubData:    make(chan subData, 1000),
		closeSub:       make(chan struct{}, 1),
		haveMiddleware: false,
	}

	go store.watchSub()

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

func (s *SyncStore) Close() {
	s.closeSub <- struct{}{}
}

func (s *SyncStore) dispatch(action Action) {
	if s.haveMiddleware {
		for _, handler := range s.middlewares {
			handler(s.dispatch, action, s.next)
		}
	} else {
		s.next(action)
	}
}

func (s *SyncStore) watchSub() {
	for {
		select {
		case <-s.closeSub:
			return
		case sd := <-s.chanSubData:
			for _, handler := range s.subscibers {
				handler(sd.state, sd.action)
			}
		}
	}
}

func (s *SyncStore) next(action Action) {
	s.locker.Lock()
	for _, handler := range s.reducers {
		s.state = handler(s.state, action)
	}
	s.locker.Unlock()
	s.chanSubData <- subData{action, s.state}
}
