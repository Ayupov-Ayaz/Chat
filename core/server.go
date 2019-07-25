package core

type Server struct {
	// Зарегистрированные клиенты
	clients map[*client]bool

	// входящие сообщения от клиентов
	broadcast chan []byte

	// Подписка от клиентов
	subscribe chan *client

	// Отписка от клиента
	unsubscribe chan *client
}

func NewServer() *Server {
	return &Server{
		broadcast: make(chan []byte),
		subscribe: make(chan *client),
		unsubscribe: make(chan *client),
		clients: make(map[*client] bool),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <- s.subscribe:
			s.clients[client] = true
		case client := <- s.unsubscribe:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				// закрываем канал отправки сообщений
				close(client.send)
			}

		case message := <- s.broadcast:
			for client := range s.clients {
				select {
				// рассылка сообщений
				case client.send <- message:
				default:
					// websocket закрыт
					close(client.send)
					delete(s.clients, client)
				}
			}

		}
	}
}