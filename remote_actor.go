package actor

import (
    "context"
    "encoding/json"
    "fmt"
    "net"
    "sync"
    "sync/atomic"
)

// remoteActorRef 表示远程actor的引用
type remoteActorRef struct {
    id      string
    address string
    conn    net.Conn
    mu      sync.RWMutex
    msgId   atomic.Int64
}
type remoteMessage struct {
    Id      int64
    Type    string
    Target  string
    Payload []byte
    Error   string
}

func (r *remoteActorRef) Address() string {
    return r.address
}

// NewRemoteActorRef 创建一个远程actor引用
func NewRemoteActorRef(id string, address string) (ActorRef, error) {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return nil, err
    }
    
    return &remoteActorRef{
        id:      id,
        address: address,
        conn:    conn,
    }, nil
}

func (r *remoteActorRef) Request(ctx context.Context, req interface{}, resp interface{}) error {
    // 序列化请求
    data, err := json.Marshal(req)
    if err != nil {
        return err
    }
    
    // 发送请求
    r.mu.RLock()
    conn := r.conn
    r.mu.RUnlock()
    
    if conn == nil {
        return fmt.Errorf("connection is closed")
    }
    
    msg := remoteMessage{
        Id:      r.msgId.Add(1),
        Target:  r.id,
        Type:    "Request",
        Payload: data,
    }
    msgData, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    
    _, err = conn.Write(msgData)
    if err != nil {
        r.mu.Lock()
        r.conn = nil
        r.mu.Unlock()
        return err
    }
    
    // 读取响应
    var replyMsg remoteMessage
    dec := json.NewDecoder(conn)
    if err := dec.Decode(&replyMsg); err != nil {
        return err
    }
    if replyMsg.Error != "" {
        return fmt.Errorf(replyMsg.Error)
    }
    // 赋值数据到resp
    return json.Unmarshal(replyMsg.Payload, resp)
}

func (r *remoteActorRef) Tell(msg interface{}) error {
    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    
    r.mu.RLock()
    conn := r.conn
    r.mu.RUnlock()
    
    if conn == nil {
        return fmt.Errorf("connection is closed")
    }
    
    message := remoteMessage{
        Id:      r.msgId.Add(1),
        Target:  r.id,
        Type:    "Tell",
        Payload: data,
    }
    msgData, err := json.Marshal(message)
    if err != nil {
        return err
    }
    
    _, err = conn.Write(msgData)
    if err != nil {
        r.mu.Lock()
        r.conn = nil
        r.mu.Unlock()
    }
    return err
}

func (r *remoteActorRef) ID() string {
    return r.id
}

func (r *remoteActorRef) onResponse(id int64, payload []byte, err string) {
    fmt.Println("onResponse:", id, string(payload), err)
}
