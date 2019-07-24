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
