package godux

type dispatcher struct {
	events chan Action
}

func NewDispatcher() dispatcher {
	return dispatcher{
		events: make(chan Action, 1000),
	}
}

func (d *dispatcher) Dispatch(action Action) {
	d.events <- action
}

func (d *dispatcher) GetAction() Action {
	return <-d.events
}
