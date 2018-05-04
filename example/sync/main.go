package main

import (
	"fmt"
	"time"

	"github.com/plimble/godux"
)

const (
	GET_USER         = "GET_USER"
	INCREASE_COUNTER = "INCREASE_COUNTER"
)

type UserState struct {
	Username string
}

type AppState struct {
	Count int
	User  UserState
}

func GetUser(username string) godux.ActionCreator {
	return func(dispatch godux.Dispatch, getState godux.GetState) {
		dispatch(godux.Action{
			Type:    GET_USER,
			Payload: username,
		})
	}
}

func IncreaseCounter() godux.ActionCreator {
	return func(dispatch godux.Dispatch, getState godux.GetState) {
		dispatch(godux.Action{
			Type: INCREASE_COUNTER,
		})
	}
}

func UserReducer(state interface{}, action godux.Action) interface{} {
	appState := state.(AppState)
	switch action.Type {
	case GET_USER:
		appState.User.Username = action.Payload.(string)
		return appState
	default:
		return state
	}
}

func CounterReducer(state interface{}, action godux.Action) interface{} {
	appState := state.(AppState)
	switch action.Type {
	case INCREASE_COUNTER:
		appState.Count += 1
		return appState
	default:
		return state
	}
}

func Logger(dispatch godux.Dispatch, action godux.Action, next godux.Next) {
	fmt.Println("Before")
	next(action)
	fmt.Println("After")
}

func main() {
	store := godux.NewSyncStore(AppState{}, []godux.ReducerHandler{
		UserReducer,
		CounterReducer,
	})
	defer store.Close()

	store.ApplyMiddleware(Logger)

	store.Subscribe(func(state interface{}, action godux.Action) {
		s := state.(AppState)
		fmt.Println("Action", action.Type, "Count", s.Count, "Username", s.User.Username)
	})

	store.Dispatch(GetUser("userABC"))
	store.Dispatch(IncreaseCounter())
	store.Dispatch(IncreaseCounter())
	store.Dispatch(IncreaseCounter())
	store.Dispatch(GetUser("userXYZ"))
	time.Sleep(1 * time.Second)
}
