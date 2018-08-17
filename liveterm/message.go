package liveterm

type Message struct {
    MsgType  string
    Username string
    Password string
}

type AuthMessage struct {
    Username string
    Password string
}

type SessionMessage struct {
    Cmd      string
    Arg      string   // listの時に指定すればそのセッションの詳細情報、なければセッションのリストが帰る
}

type Response struct {
    Result bool
    Msg    string
}
