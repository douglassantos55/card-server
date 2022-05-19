package server

type Handler interface {
	Process(event Event, dispatcher *Dispatcher)
}

type Dispatcher struct {
	handlers []Handler

	Register chan Handler
	Dispatch chan Event
}

func NewDispatcher() *Dispatcher {
	dispatcher := &Dispatcher{
		handlers: make([]Handler, 0),

		Register: make(chan Handler),
		Dispatch: make(chan Event),
	}

	go func() {
		for {
			select {
			case handler := <-dispatcher.Register:
				dispatcher.handlers = append(dispatcher.handlers, handler)
			case event := <-dispatcher.Dispatch:
				for _, handler := range dispatcher.handlers {
					handler.Process(event, dispatcher)
				}
			}
		}
	}()

	return dispatcher
}
