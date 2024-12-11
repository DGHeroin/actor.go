package actor

import (
    "context"
    "encoding/json"
    "fmt"
    "net"
    "sync"
    "sync/atomic"
    "time"
)

// remoteActorRef 表示远程actor的引用
type remoteActorRef struct {
    id        string
    address   string
    conn      net.Conn
    mu        sync.RWMutex
    msgId     atomic.Int64
    pending   map[int64]chan Response
    pendingMu sync.RWMutex
    writeCh   chan interface{}
}

// NewRemoteActorRef 创建一个远程actor引用
func NewRemoteActorRef(id string, address string) (ActorRef, error) {
    conn, err := net.DialTimeout("tcp", address, time.Second*5)
    if err != nil {
        return nil, fmt.Errorf("connect failed: %w", err)
    }
    
    ref := &remoteActorRef{
        id:      id,
        address: address,
        conn:    conn,
        pending: make(map[int64]chan Response),
        writeCh: make(chan interface{}, 100), // 添加缓冲区以避免阻塞
    }
    
    go ref.readLoop()
    go ref.writeLoop()
    
    return ref, nil
}

// readLoop 持续读取响应
func (r *remoteActorRef) readLoop() {
    defer r.cleanup()
    
    dec := json.NewDecoder(r.conn)
    for {
        var msg struct {
            Id      int64       `json:"id"`
            Error   string      `json:"error,omitempty"`
            Payload interface{} `json:"payload,omitempty"`
        }
        
        if err := dec.Decode(&msg); err != nil {
            r.handleError(fmt.Errorf("read failed: %w", err))
            return
        }
        
        r.pendingMu.RLock()
        ch, ok := r.pending[msg.Id]
        r.pendingMu.RUnlock()
        
        if ok {
            ch <- Response{Data: msg.Payload, Error: msg.Error}
            r.pendingMu.Lock()
            delete(r.pending, msg.Id)
            close(ch)
            r.pendingMu.Unlock()
        }
    }
}

// writeLoop 持续写入请求
func (r *remoteActorRef) writeLoop() {
    enc := json.NewEncoder(r.conn)
    for msg := range r.writeCh {
        if err := enc.Encode(msg); err != nil {
            r.handleError(fmt.Errorf("write failed: %w", err))
            return
        }
    }
}

// cleanup 清理资源
func (r *remoteActorRef) cleanup() {
    r.mu.Lock()
    if r.conn != nil {
        _ = r.conn.Close()
        r.conn = nil
    }
    r.mu.Unlock()
    
    // 通知所有等待的请求
    r.pendingMu.Lock()
    for id, ch := range r.pending {
        ch <- Response{Error: "connection closed"}
        close(ch)
        delete(r.pending, id)
    }
    r.pendingMu.Unlock()
}

// handleError 处理错误
func (r *remoteActorRef) handleError(err error) {
    r.cleanup()
    fmt.Println(err)
}

// Request 发送请求并等待响应
func (r *remoteActorRef) Request(ctx context.Context, req interface{}, resp interface{}) error {
    r.mu.RLock()
    conn := r.conn
    r.mu.RUnlock()
    
    if conn == nil {
        return fmt.Errorf("connection is closed")
    }
    
    msgID := r.msgId.Add(1)
    respCh := make(chan Response, 1)
    
    r.pendingMu.Lock()
    r.pending[msgID] = respCh
    r.pendingMu.Unlock()
    
    msg := struct {
        Id      int64       `json:"id"`
        Target  string      `json:"target"`
        Type    string      `json:"type"`
        Payload interface{} `json:"payload"`
    }{
        Id:      msgID,
        Target:  r.id,
        Type:    "Request",
        Payload: req,
    }
    
    r.writeCh <- msg
    
    select {
    case <-ctx.Done():
        r.pendingMu.Lock()
        delete(r.pending, msgID)
        r.pendingMu.Unlock()
        return ctx.Err()
    case response := <-respCh:
        if response.Error != "" {
            return fmt.Errorf(response.Error)
        }
        if resp != nil {
            data, err := json.Marshal(response.Data)
            if err != nil {
                return fmt.Errorf("marshal response failed: %w", err)
            }
            return json.Unmarshal(data, resp)
        }
        return nil
    }
}

// Tell 发送单向消息
func (r *remoteActorRef) Tell(msg interface{}) error {
    r.mu.RLock()
    conn := r.conn
    r.mu.RUnlock()
    
    if conn == nil {
        return fmt.Errorf("connection is closed")
    }
    
    m := struct {
        Target  string      `json:"target"`
        Type    string      `json:"type"`
        Payload interface{} `json:"payload"`
    }{
        Target:  r.id,
        Type:    "Tell",
        Payload: msg,
    }
    
    r.writeCh <- m
    
    return nil
}

func (r *remoteActorRef) ID() string {
    return r.id
}

func (r *remoteActorRef) Address() string {
    return r.address
}
