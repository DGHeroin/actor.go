package actor

import (
    "context"
    "errors"
)

var (
    ErrActorNotFound = errors.New("actor not found")
    ErrMailboxFull   = errors.New("actor mailbox is full")
    ErrActorStopped  = errors.New("actor is stopped")
)

// MessageType 定义消息类型
type MessageType string

const (
    MessageTypeRequest MessageType = "Request"
    MessageTypeTell    MessageType = "Tell"
)

// Message 表示一个actor消息
type Message struct {
    Payload interface{}     // 消息内容
    ReplyTo chan Response   `json:"-"` // 响应通道
    Context context.Context `json:"-"` // 消息上下文
}

// Response 表示一个响应
type Response struct {
    Data  interface{} // 响应数据
    Error string      // 错误信息
}

// remoteMessage 表示一个远程消息
type remoteMessage struct {
    Type    MessageType `json:"type"`
    Id      int64       `json:"id"`
    Target  string      `json:"target"`
    Payload interface{} `json:"payload"`
    Error   string      `json:"error,omitempty"`
}

// ActorRef 表示actor的引用
type ActorRef interface {
    // Request 发送请求并等待响应
    Request(ctx context.Context, req interface{}, resp interface{}) error
    
    // Tell 发送单向消息，不等待响应
    Tell(msg interface{}) error
    
    // ID 获取Actor的ID
    ID() string
    
    // Address 获取Actor的地址
    Address() string
}
