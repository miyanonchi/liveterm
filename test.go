package main

import (
    "fmt"
    "os"
    "os/exec"
    "io"
    "net/http"

    "golang.org/x/net/websocket"
    "github.com/kr/pty"
)

func startSession() (*os.File, error) {
    // Create arbitrary command.
    // c := exec.Command("script", "--timing=session.timing.log", "-qf", "session.log")

    fmt.Println("run test.sh")
    cmd := exec.Command("./test.sh")

    // Start the command with a pty.
    ptmx, err := pty.Start(cmd)
    if err != nil {
        return nil, err
    }

    return ptmx, nil
}

func echoHandler(ws *websocket.Conn) {
    ptmx, err := startSession()

    if err != nil {
        panic(err)
    }

    println("get ptmx")

    ws.Write(([]byte)("hello"))

    //// Handle pty size.
    //ch := make(chan os.Signal, 1)
    //signal.Notify(ch, syscall.SIGWINCH)
    //go func() {
    //    for range ch {
    //       if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
    //           log.Printf("error resizing pty: %s", err)
    //       }
    //    }
    //}()
    //ch <- syscall.SIGWINCH // Initial resize.

    //fmt.Println("syscall.SIGWINCH")

    //// Set stdin in raw mode.
    //oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
    //if err != nil {
    //    return nil, err
    //}
    //defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

    // Copy stdin to the pty and the pty to stdout.
    //go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
    //_, _ = io.Copy(os.Stdout, ptmx)

    doneCh := make(chan bool, 1)
    fmt.Println("in goroutine")

    // Make sure to close the pty at the end.
    defer func() {
        println("ptmx close")
        _ = ptmx.Close()
    }() // Best effort.

    for {
      select {
      // receive done request
      case <-doneCh:
        fmt.Println("done")
        ws.Close()
        return

      // read data from websocket connection
      default:
        var msg []byte
        size, err := ws.Read(msg)
        if err == io.EOF {
          print("EOF!")
          doneCh <- true
        } else if err != nil {
          fmt.Printf("%+v\n", ws)
          doneCh <- true
        } else if size != 0 {
          ptmx.Write(msg)
          fmt.Println(msg)
        }
      }
    }

    fmt.Println("last")
}

func main() {
    http.Handle("/ws", websocket.Handler(echoHandler))
    err := http.ListenAndServe("192.168.0.128:8080", nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
