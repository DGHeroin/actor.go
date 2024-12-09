package actor

import "context"

// Message 表示actor之间传递的消息
type Message struct {
    Payload interface{}     // 消息内容
    ReplyTo chan Response   `json:"-"` // 响应通道
    Context context.Context `json:"-"` // 消息上下文
}

// Response 表示请求的响应
type Response struct {
    Data  interface{} // 响应数据
    Error string      // 错误信息
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
