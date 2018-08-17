package liveterm

import (
    "errors"
    "os"
    "os/exec"
    "os/signal"
    "strconv"
    "syscall"
    "io"
    "log"
    "time"

    "github.com/kr/pty"
    "github.com/gorilla/websocket"
)

type Session struct {
    id          uint64
    server      *Server
    name        string
    broadcast   chan []byte
    owner       *Client
    watchers    []*Client
    startedAt   time.Time
    finishedAt  time.Time
    isStarted   bool
    logDir      string
    cmd         []string
}

// ユニークなIDを作る
var uniqId = func() (func() uint64) {
      seqCh := make(chan uint64, 1)

      go func() {
          for i := uint64(0); ; i++ {
             seqCh <- i
          }
      }()

      return func() uint64 {
          now := time.Now()

          var id uint64
          id = 0
          id +=            uint64(now.Year())
          id += id * 100 + uint64(now.Month())
          id += id * 100 + uint64(now.Day())
          id += id * 100 + uint64(now.Hour())
          id += id * 100 + uint64(now.Minute())
          id += id * 100 + uint64(now.Second())
          id += id * 1000 + <- seqCh

          return id
      }
}()

func NewSession(s *Server, name string, owner *Client) *Session {
    sid        := uniqId()
    watchers   := []*Client{}
    logDir     := "/sessions/"

    timingFile := logDir + "session." + strconv.FormatUint(sid, 10) + ".timing.log"
    logFile    := logDir + "session." + strconv.FormatUint(sid, 10) + ".log"
    cmd        := []string{"script", "--timing=" + timingFile, "-qf", logFile}

    return &Session{
        id:        sid,
        server :   s,
        name:      name,
        broadcast: make(chan []byte),
        owner:     owner,
        watchers:  watchers,
        isStarted: false,
        logDir:    logDir,
        cmd:       cmd,
    }
}

func (s *Session) Start() error {
    if s.logDir == "" {
        return errors.New("no log dir")
    }

    if _, err := os.Stat(s.logDir); err != nil {
        return errors.New(s.logDir + " not exist")
    }

    s.startedAt = time.Now()
    s.isStarted = true

    s.startSession()

    s.owner.Send([]byte("Finished"))

    return nil
}

/*
 * ws, session_id
 */
func (s *Session) startSession() error {
    cmd := exec.Command(s.cmd[0], s.cmd[1:]...)

    log.Println("created cmd")

    // Start the command with a pty.
    ptmx, err := pty.Start(cmd)
    if err != nil {
        return err
    }

    log.Println("started pty")

    // セッション終了
    defer func() {
        _ = ptmx.Close()

        log.Println("closed pty")

        s.finishedAt = time.Now()
        s.isStarted = false
    }()

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

    // Copy stdin to the pty and the pty to stdout.
    go func() {
        for {
            _, w, err := s.owner.ws.NextReader()

            if err != nil {
                log.Println(err)
                return
            }

            _, _ = io.Copy(ptmx, w)
            log.Println("copy 1")
        }
    }()


    go func() {
        for {
            r, err := s.owner.ws.NextWriter(websocket.BinaryMessage)
            if err != nil {
                log.Println(err)
                return
            }

            _, _ = io.Copy(r, ptmx)
            log.Println("copy 2")
        }
    }()

    //go func() {
    //    var buf []byte
    //    for {
    //        cnt, err := ptmx.Read(buf)

    //        if err != nil {
    //            // log.Println(err)
    //            break;
    //        }

    //        s.owner.ws.WriteMessage(websocket.BinaryMessage, buf)
    //    }
    //}()

    //go func() {
    //    for {
    //        var buf []byte
    //        _, buf, err := s.owner.ws.ReadMessage()

    //        if err != nil {
    //            // log.Println(err)
    //            break;
    //        }

    //        // 改行を追加
    //        buf = append(buf, 0x0d)
    //        ptmx.Write(buf)
    //    }
    //}()

    cmd.Wait()

    return nil
}

func (s *Session) watch(client *Client) {
}
