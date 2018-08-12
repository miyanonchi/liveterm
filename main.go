package main

import (
  "log"

  "./liveterm"
)

func main() {

    liveterm.NewServer("/ws")

    go liveterm.Listen()
}
