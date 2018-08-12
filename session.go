package liveterm

import (
    "os"
    "os/exec"
    "os/signal"
    "syscall"
    "io"
    "log"
    "time"

    "golang.org/x/net/websocket"
    "github.com/kr/pty"
    "golang.org/x/crypto/ssh/terminal"
)

type Session struct {
    id          int
    server      Server
    name        string
    owners      []*Client

    watchers    []*Client
    startedAt  Time
    finishedAt Time
    is_started  bool
    logDir     string
    cmd         Cmd
    ptmx        *os.File
}

func (s *Session) Start() error {
    if len(owners) == 0 {
        return errors.New("there is no owner")
    }

    if logDir == "" {
        return errors.New("no log dir")
    }

    if _, err := os.Stat(logDir); err != nil {
        return errors.New("log dir not exists")
    }

    s.startedAt = time.Now()
    s.is_started = true

    s.startSession()
}

func (s *Session) End() error {
    s.finishedAt = time.Now()
    s.is_started = false
}

func (s *Session) AddOwner(client *Client) {
  append(s.watchers, client)
}

/*
 * ws, session_id
 */
func (s *Session) startSession() error {
    // Create arbitrary command.
    // c := exec.Command("script", "--timing=session.timing.log", "-qf", "session.log")

    time := s.startedAt.Format("20000101000000")
    opt1 := "--timing="
    opt1 += s.logDir + "/session." + time + ".timing.log"
    opt2 := "-qf"
    opt3 := s.logDir + "/session." + time + ".log"

    print(time)

    cmd := exec.Command("echo", "hogehoge")

    // Start the command with a pty.
    ptmx, err := pty.Start(cmd)
    if err != nil {
        return err
    }

    // Make sure to close the pty at the end.
    defer func() { _ = ptmx.Close() }() // Best effort.

    s.ptmx = ptmx

    // Handle pty size.
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGWINCH)
    go func() {
        for range ch {
           if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
               log.Printf("error resizing pty: %s", err)
           }
        }
    }()
    ch <- syscall.SIGWINCH // Initial resize.

    // Set stdin in raw mode.
    oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        retrun err
    }
    defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

    // Copy stdin to the pty and the pty to stdout.
    go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
    _, _ = io.Copy(os.Stdout, ptmx)

    return nil
}

func (s *Session) watch(client *Client) {
}
