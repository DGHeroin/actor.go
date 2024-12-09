package actor

import (
    "context"
    "encoding/json"
    "fmt"
    "net"
)

type ActorServer struct {
    listener net.Listener
    sys      *ActorSystem
}

func NewActorServer(sys *ActorSystem) *ActorServer {
    s := &ActorServer{
        sys: sys,
    }
    return s
}

func (s *ActorServer) Serve(addr string) error {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    
    // 启动处理连接的goroutine
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                return
            }
            go s.handleConnection(conn)
        }
    }()
    
    return nil
}

func (s *ActorServer) handleConnection(conn net.Conn) {
    defer conn.Close()
    ctx := context.Background()
    
    dec := json.NewDecoder(conn)
    enc := json.NewEncoder(conn)
    responseCh := make(chan *remoteMessage, 100)
    go func() {
        for {
            select {
            case msg := <-responseCh:
                err := enc.Encode(msg)
                if err != nil {
                    return
                }
            }
        }
    }()
    for {
        select {
        
        default:
            var msg remoteMessage
            err := dec.Decode(&msg)
            if err != nil {
                return
            }
            
            switch msg.Type {
            case "Request":
                go s.handleRequestMessage(ctx, msg, responseCh)
            case "Tell":
                if err := s.sys.SendMessage(msg.Target, Message{
                    Payload: msg.Payload,
                    Context: ctx,
                }); err != nil {
                    fmt.Println(err)
                    return
                }
            case "Response":
                s.sys.onResponse(msg.Target, msg.Id, msg.Payload, msg.Error)
            default:
                fmt.Println("unknown message type:", msg.Type)
            }
        }
    }
}

func (s *ActorServer) handleRequestMessage(ctx context.Context, msg remoteMessage, rch chan *remoteMessage) {
    ch := make(chan Response, 1)
    defer close(ch)
    // 传递消息给actor
    err := s.sys.SendMessage(msg.Target, Message{
        Payload: msg.Payload,
        ReplyTo: ch,
        Context: ctx,
    })
    if err != nil {
        rch <- &remoteMessage{
            Type:   "Response",
            Id:     msg.Id,
            Target: msg.Target,
            Error:  err.Error(),
        }
        return
    }
    // 拿到消息
    replyMsg := <-ch
    var replyData []byte
    var replyError string
    switch v := replyMsg.Data.(type) {
    case interface{}:
        if d, err := json.Marshal(v); err == nil {
            replyData = d
        } else {
            replyError = err.Error()
        }
    case []byte, string:
        replyData = replyMsg.Data.([]byte)
    default:
        replyError = "invalid response type"
    }
    rch <- &remoteMessage{
        Type:    "Response",
        Id:      msg.Id,
        Target:  msg.Target,
        Payload: replyData,
        Error:   replyError,
    }
}
