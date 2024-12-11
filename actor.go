package actor

import (
    "sync/atomic"
)

// Actor 表示一个基本的actor
type Actor struct {
    id       string
    mailbox  chan Message
    handler  MessageHandler
    done     chan struct{}
    stopping atomic.Bool // 使用原子操作标记停止状态
}

type MessageHandler func(msg interface{}) (interface{}, error)

// NewActor 创建一个新的actor
func NewActor(id string, handler MessageHandler) *Actor {
    return &Actor{
        id:      id,
        mailbox: make(chan Message, 100),
        handler: handler,
        done:    make(chan struct{}),
    }
}

// Start 启动actor的消息处理循环
func (a *Actor) Start() {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                // 处理 panic，可以选择重启 actor
                close(a.done)
            }
        }()
        defer close(a.done)
        
        for msg := range a.mailbox {
            if a.stopping.Load() {
                // Actor 正在停止，拒绝新消息
                if msg.ReplyTo != nil {
                    msg.ReplyTo <- Response{Error: ErrActorStopped.Error()}
                }
                continue
            }
            
            a.handleMessage(msg)
        }
    }()
}

// handleMessage 处理单个消息
func (a *Actor) handleMessage(msg Message) {
    if msg.Context != nil {
        select {
        case <-msg.Context.Done():
            if msg.ReplyTo != nil {
                msg.ReplyTo <- Response{Error: msg.Context.Err().Error()}
            }
            return
        default:
        }
    }
    
    result, err := a.handler(msg.Payload)
    if msg.ReplyTo == nil {
        return // 单向消息，忽略结果
    }
    
    if err != nil {
        msg.ReplyTo <- Response{Error: err.Error()}
        return
    }
    
    msg.ReplyTo <- Response{Data: result}
}

// Stop 停止actor
func (a *Actor) Stop() {
    if !a.stopping.CompareAndSwap(false, true) {
        return // 已经在停止过程中
    }
    
    close(a.mailbox)
    <-a.done // 等待所有消息处理完成
}

// Send 发送消息给actor
func (a *Actor) Send(msg Message) error {
    if a.stopping.Load() {
        return ErrActorStopped
    }
    
    if msg.Context != nil {
        select {
        case <-msg.Context.Done():
            if msg.ReplyTo != nil {
                msg.ReplyTo <- Response{Error: msg.Context.Err().Error()}
            }
            return msg.Context.Err()
        case a.mailbox <- msg:
            return nil
        default:
            return ErrMailboxFull
        }
    }
    
    select {
    case a.mailbox <- msg:
        return nil
    default:
        return ErrMailboxFull
    }
}

// IsStopped 检查actor是否已停止
func (a *Actor) IsStopped() bool {
    return a.stopping.Load()
}

// ToRef 获取对该Actor的本地引用
func (a *Actor) ToRef() ActorRef {
    return &localActorRef{actor: a, id: a.id}
}
