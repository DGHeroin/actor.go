package actor

// Actor 表示一个基本的actor
type Actor struct {
    id      string
    mailbox chan Message
    handler MessageHandler
}
type MessageHandler func(msg interface{}) (interface{}, error)

// NewActor 创建一个新的actor
func NewActor(id string, handler MessageHandler) *Actor {
    return &Actor{
        id:      id,
        mailbox: make(chan Message, 100),
        handler: handler,
    }
}

// Start 启动actor的消息处理循环
func (a *Actor) Start() {
    go func() {
        for msg := range a.mailbox {
            if msg.ReplyTo == nil {
                // 单向消息，不需要响应
                _, _ = a.handler(msg.Payload)
                continue
            }
            
            // 处理请求消息
            result, err := a.handler(msg.Payload)
            if err != nil {
                msg.ReplyTo <- Response{Error: err.Error()}
                continue
            }
            
            msg.ReplyTo <- Response{Data: result}
        }
    }()
}

// Stop 停止actor
func (a *Actor) Stop() {
    close(a.mailbox)
}

// Send 发送消息给actor
func (a *Actor) Send(msg Message) error {
    select {
    case <-msg.Context.Done():
        if msg.ReplyTo != nil {
            msg.ReplyTo <- Response{Error: msg.Context.Err().Error()}
        }
        return msg.Context.Err()
    case a.mailbox <- msg:
        return nil
    }
}

// ToRef 获取对该Actor的本地引用
func (a *Actor) ToRef() ActorRef {
    return &localActorRef{actor: a, id: a.id}
}
