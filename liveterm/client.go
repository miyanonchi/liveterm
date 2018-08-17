package liveterm

import (
    "fmt"
    "log"
    "encoding/json"
    "github.com/gorilla/websocket"
)

type Client struct {
    id              uint64
    server          *Server
    ws              *websocket.Conn
    isAuthenticated bool
    username        string
    session         *Session
}

// Create new chat server.
func NewClient(s *Server, ws *websocket.Conn) *Client {
    id := uniqId()

    return &Client{
        id: id,
        server: s,
        ws: ws,
        isAuthenticated: false,
    }
}

func (c *Client) Send(bytes []byte) {
     c.ws.WriteMessage(websocket.TextMessage, bytes)
}

func (c *Client) SendResponse(res *Response) {
     bytes, err := json.Marshal(res)

     if err != nil {
         log.Println("Response marshal error:", err)
     }

     c.Send(bytes)
}

func (c *Client) Listen() {
    for {
        _, message, err := c.ws.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            return
        }

        fmt.Printf("message: '%s'\n", string(message))

        msg := new(Message)
        if err := json.Unmarshal(message, &msg); err != nil {
            c.SendResponse(&Response{ false, "JSON unmarshal error"})
            log.Println("JSON Unmarshal error:", err)
            log.Println("input:", string(message))
            return
        }

        if msg.MsgType == "auth" {
            c.Auth(message)
        }

        if msg.MsgType == "session" {
            c.Session(message)
        }
    }
}

func (c *Client) Auth(msg []byte) {
     var auth AuthMessage

     if err := json.Unmarshal(msg, &auth); err != nil {
         c.SendResponse(&Response{ false, "JSON unmarshal error"})
         log.Println("JSON Unmarshal error:", err)
     }

    if auth.Username == auth.Password {
        c.isAuthenticated = true
        c.username = auth.Username

        c.SendResponse(&Response{true, "OK"})
    } else {
        c.SendResponse(&Response{false, ""})
    }
}

func (c *Client) Session(message []byte) {
    var msg SessionMessage

     if err := json.Unmarshal(message, &msg); err != nil {
         c.SendResponse(&Response{ false, "JSON unmarshal error"})
         log.Println("JSON Unmarshal error:", err)
     }

     if msg.Cmd == "list" {
         // do something
         bytes, err := json.Marshal(c.server.sessions)

         if err != nil {
             c.SendResponse(&Response{ false, "Session marshal error"})
             log.Println("Session marshal error:", err)
             return
         }

         c.Send(bytes)
     }

     if msg.Cmd == "create" || msg.Cmd == "start" || msg.Cmd == "delete"{
         //if !c.isAuthenticated {
         //   c.SendResponse(&Response{ false, "Authenticate before start/stop msg"})
         //}

         if msg.Arg == "" {
            c.SendResponse(&Response{false, "Arg is not allowd empty value"})
         }

         session := NewSession(c.server, msg.Arg, c)

         c.session = session
         c.server.SessionAdd(session)

         if err := session.Start(); err != nil {
             c.SendResponse(&Response{false, err.Error()})
             log.Println("Session start failed:", err)
             return
         }

         c.SendResponse(&Response{true, "OK"})
     }
}
