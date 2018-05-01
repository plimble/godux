package godux

type ActionCreator func(dispatch Dispatch, getState GetState)

type Action struct {
	Type    string
	Payload interface{}
	Meta    interface{}
	Error   bool
}
