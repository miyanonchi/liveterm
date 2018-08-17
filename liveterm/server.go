package liveterm

import (
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

type Server struct {
    pattern      string
    clients      map[uint64]*Client
    sessions     map[uint64]*Session
    clientAddCh  chan *Client
    clientDelCh  chan *Client
    sessionAddCh chan *Session
    sessionDelCh chan *Session
    doneCh       chan bool
    errCh        chan error
}

func NewServer(pattern string) *Server {
    clients      := make(map[uint64]*Client)
    sessions     := make(map[uint64]*Session)
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

// Listen and serve.
// It serves client connection and broadcast request.
func (s *Server) Listen() {
    // handle ws connection
    http.Handle(s.pattern, http.HandlerFunc(s.handleRequest))

    // handle static contents
    http.Handle("/", http.FileServer(http.Dir("../public")))

    // start server main loop
    go func() {
        err := http.ListenAndServe(":8080", nil)
        if err != nil {
            log.Fatal("ListenAndServe: ", err)
        }
    }()

    log.Println("Listening...")

    for {
        select {

        // add new client
        case c := <-s.clientAddCh:
            log.Println("add new client:", c.id)
            s.clients[c.id] = c

        // delete a client
        case c := <-s.clientDelCh:
            log.Println("delete a client:", c.id)
            delete(s.clients, c.id)

        // add new session
        case c := <-s.sessionAddCh:
            log.Println("add new session:", c.id)
            s.sessions[c.id] = c

        // delete a session
        case c := <-s.sessionDelCh:
            log.Println("delete session:", c.id)
            delete(s.sessions, c.id)

        case err := <-s.errCh:
            log.Println("Error:", err.Error())

        case <-s.doneCh:
            log.Println("Shutdown server.")

            //// 記録中のセッションを全て保存する
            //for id, session := range s.sessoins {
            //    session.Finish()
            //}

            return
        }
    }
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
    // Upgrade initial GET request to a websocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Make sure we close the connection when the function returns
    defer ws.Close()

    client := NewClient(s, ws)

    s.clientAddCh <- client

    client.Listen()
}

func (s *Server) SessionAdd(c *Session) {
    s.sessionAddCh <- c
}

func (s *Server) SessionDel(c *Session) {
    s.sessionDelCh <- c
}

func (s *Server) ClientAdd(c *Client) {
    log.Println("add new client [", c.id, "]")
    log.Println("now ", len(s.clients), " clients connected")
    s.clientAddCh <- c
}

func (s *Server) ClientDel(c *Client) {
    log.Println("delete client [", c.id, "]")
    log.Println("now ", len(s.clients), " clients connected")
    s.clientDelCh <- c
}

func (s *Server) Shutdown() {
    s.doneCh <- true
}

func (s *Server) Err(err error) {
    s.errCh <- err
}
