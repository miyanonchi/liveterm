package liveterm

import (
    "log"
    "net/http"

    "golang.org/x/net/websocket"
)

// Chat server.
type Server struct {
    pattern      string
    clients        map[int]*Client
    sessions     map[int]*Session
    clientAddCh    chan *Client
    clientDelCh    chan *Client
    sessionAddCh chan *Session
    sessionDelCh chan *Session
    doneCh       chan bool
    errCh        chan error
}

// Create new chat server.
func NewServer(pattern string) *Server {
    clients      := make(map[int]*Client)
    sessions     := make(map[int]*Session)
    clientAddCh  := make(chan *Client)
    clientDelCh  := make(chan *Client)
    sessionAddCh := make(chan *Session)
    sessionDelCh := make(chan *Session)
    doneCh       := make(chan bool)
    errCh        := make(chan error)

    return &Server{
        pattern,
        clients,
        sessions,
        clientAddCh,
        clientDelCh,
        sessionAddCh,
        sessionDelCh,
        doneCh,
        errCh,
    }
}

func (s *Server) SessionAdd(c *Session) {
    s.sessionAddCh <- c
}

func (s *Server) SessionDel(c *Session) {
    s.sessionDelCh <- c
}

func (s *Server) ClientAdd(c *Client) {
    s.clientAddCh <- c
}

func (s *Server) ClientDel(c *Client) {
    s.clientDelCh <- c
}

func (s *Server) Done() {
    s.doneCh <- true
}

func (s *Server) Err(err error) {
    s.errCh <- err
}

// Listen and serve.
// It serves client connection and broadcast request.
func (s *Server) Listen() {

    log.Println("Listening server...")

    // websocket handler
    onConnected := func(ws *websocket.Conn) {
        defer func() {
            err := ws.Close()
            if err != nil {
                s.errCh <- err
            }
        }()

        client := NewClient(ws, s)
        s.Add(client)
        client.Listen()
    }
    http.Handle(s.pattern, websocket.Handler(onConnected))
    log.Println("Created handler")

    for {
        select {

        // Add new a client
        case c := <-s.addCh:
            log.Println("Added new client")
            s.clients[c.id] = c
            log.Println("Now", len(s.clients), "clients connected.")
            s.sendPastMessages(c)

        // del a client
        case c := <-s.delCh:
            log.Println("Delete client")
            delete(s.clients, c.id)

        // broadcast message for all clients
        case msg := <-s.sendAllCh:
            log.Println("Send all:", msg)
            //s.messages = append(s.messages, msg)
            s.sendAll(msg)

        case err := <-s.errCh:
            log.Println("Error:", err.Error())

        case <-s.doneCh:
            return
        }
    }
}
