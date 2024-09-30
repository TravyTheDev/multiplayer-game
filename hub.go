package main

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Actions),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.Register:
			if _, ok := h.Clients[cl.ID]; !ok {
				h.Clients[cl.ID] = cl
			}
		case cl := <-h.Unregister:
			delete(h.Clients, cl.ID)
			close(cl.Actions)
		case a := <-h.Broadcast:
			if _, ok := h.Clients[a.UserID]; ok {
				for _, cl := range h.Clients {
					if len(h.Clients) == 1 {
						cl.Actions <- a
					} else {
						if cl.ID != a.UserID {
							cl.Actions <- a
						}
					}
				}
			}
		}
	}
}
