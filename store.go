package godux

type SubscribeHandler func(state interface{}, action Action)
type ReducerHandler func(state interface{}, action Action) interface{}
type Dispatch func(action Action)
type MiddlewareHandler func(dispatch Dispatch, action Action, next Next)
type Next func(action Action)

type Store interface {
	GetState() interface{}
	Subscribe(handler SubscribeHandler)
	Dispatch(ActionCreator)
	ApplyMiddleware(handler MiddlewareHandler)
	Close()
}

type DefaultStore struct {
	state       interface{}
	reducers    []ReducerHandler
	subscibers  []SubscribeHandler
	middlewares []MiddlewareHandler
	dispatcher  dispatcher
	close       chan struct{}
}

func NewStore(initState interface{}, reducers []ReducerHandler) Store {
	store := &DefaultStore{
		state:       initState,
		reducers:    reducers,
		subscibers:  make([]SubscribeHandler, 0),
		dispatcher:  NewDispatcher(),
		middlewares: make([]MiddlewareHandler, 0),
		close:       make(chan struct{}, 1),
	}

	go store.watch()

	return store
}

func (s *DefaultStore) GetState() interface{} {
	return s.state
}

func (s *DefaultStore) Dispatch(actionCreator ActionCreator) {
	actionCreator(s.dispatcher.Dispatch)
}

func (s *DefaultStore) Subscribe(handler SubscribeHandler) {
	s.subscibers = append(s.subscibers, handler)
}

func (s *DefaultStore) ApplyMiddleware(handler MiddlewareHandler) {
	s.middlewares = append(s.middlewares, handler)
}

func (s *DefaultStore) Close() {
	s.close <- struct{}{}
}

func (s *DefaultStore) watch() {
	haveMiddleware := false
	if len(s.middlewares) > 0 {
		haveMiddleware = true
	}

	for {
		select {
		case <-s.close:
			return
		case action := <-s.dispatcher.events:
			if haveMiddleware {
				for _, handler := range s.middlewares {
					handler(s.dispatcher.Dispatch, action, s.next)
				}
			} else {
				s.next(action)
			}
		}
	}
}

func (s *DefaultStore) next(action Action) {
	for _, handler := range s.reducers {
		s.state = handler(s.state, action)
	}

	for _, handler := range s.subscibers {
		handler(s.state, action)
	}
}
