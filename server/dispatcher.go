package server

type Handler interface {
	Process(event Event, dispatcher *Dispatcher)
}

type Dispatcher struct {
	handlers []Handler

	Dispatch   chan Event
	Register   chan Handler
	Unregister chan Handler
}

func NewDispatcher() *Dispatcher {
	dispatcher := &Dispatcher{
		handlers: make([]Handler, 0),

		Dispatch:   make(chan Event),
		Register:   make(chan Handler),
		Unregister: make(chan Handler),
	}

	go func() {
		for {
			select {
			case handler := <-dispatcher.Register:
				dispatcher.handlers = append(dispatcher.handlers, handler)
			case handler := <-dispatcher.Unregister:
				for i, handle := range dispatcher.handlers {
					if handle == handler {
						// remove item
						dispatcher.handlers = append(
							dispatcher.handlers[:i],
							dispatcher.handlers[i+1:]...,
						)
					}
				}
			case event := <-dispatcher.Dispatch:
				for _, handler := range dispatcher.handlers {
					handler.Process(event, dispatcher)
				}
			}
		}
	}()

	return dispatcher
}

func NewTestDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make([]Handler, 0),

		Register: make(chan Handler),
		Dispatch: make(chan Event),
	}
}
