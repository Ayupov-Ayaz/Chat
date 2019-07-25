package core

type Server struct {
	// Зарегистрированные клиенты
	clients map[*Client]bool

	// входящие сообщения от клиентов
	broadcast chan []byte

	// Подписка от клиентов
	subscribe chan *Client

	// Отписка от клиента
	unsubscribe chan *Client
}

func NewServer() *Server {
	return &Server{
		broadcast: make(chan []byte),
		subscribe: make(chan *Client),
		unsubscribe: make(chan *Client),
		clients: make(map[*Client]bool),
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