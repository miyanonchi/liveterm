package main

import (
  "os"
  "os/signal"
  "syscall"
  "log"

  "./liveterm"
)

func main() {
    // サーバーオブジェクト
    server := liveterm.NewServer("/ws")

    signalCh := make(chan os.Signal, 1)

    signal.Notify(signalCh,
            syscall.SIGHUP,
            syscall.SIGINT,
            syscall.SIGTERM,
            syscall.SIGQUIT)

    // シグナルをキャッチ!
    go func() {
        for {
            s := <-signalCh

            switch s {
            // kill -SIGHUP XXXX
            case syscall.SIGHUP:
                log.Println("SIGHUP")
                server.Shutdown()
                server.Listen()
            // kill -SIGINT XXXX or Ctrl+c
            case syscall.SIGINT:
                log.Println("SIGINT")
                server.Shutdown()
            // kill -SIGTERM XXXX
            case syscall.SIGTERM:
                log.Println("SIGTERM")
                server.Shutdown()
            // kill -SIGQUIT XXXX
            case syscall.SIGQUIT:
                log.Println("SIGQUIT")
                server.Shutdown()
            }
        }
    }()

    server.Listen()
}
