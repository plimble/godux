package godux

type ActionCreator func(dispatch Dispatch, state interface{})

type Action struct {
	Type    string
	Payload interface{}
	Meta    interface{}
	Error   bool
}
