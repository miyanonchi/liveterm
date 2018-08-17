package main

import (
    "net/url"
    "io"
    "os"
    "log"
    "github.com/gorilla/websocket"
)

func main() {
    u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}

    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    defer c.Close()

    go func() {
        for {
            _, r, err := c.NextReader()

            if err != nil {
                log.Fatal(err)
            }

            io.Copy(os.Stdout, r)
        }
    }()

    for {
        w, err := c.NextWriter(websocket.BinaryMessage)

        if err != nil {
            log.Fatal(err)
        }

        io.Copy(w, os.Stdin)
    }
}

