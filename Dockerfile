FROM golang:latest

RUN useradd -u 501 -M -d /liveterm takumi && \
    go get github.com/gorilla/websocket && \
    go get github.com/kr/pty && \
    go get golang.org/x/crypto/ssh/terminal && \
    go get golang.org/x/net/websocket && \
    go get github.com/spf13/viper

USER takumi
EXPOSE 8080
CMD ["go", "run", "/liveterm/main.go"]

