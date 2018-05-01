package godux

type ActionCreator func(dispatch Dispatch)

type Action struct {
	Type    string
	Payload interface{}
	Meta    interface{}
	Error   bool
}
