package actor

import (
    "fmt"
    "sync"
)

// ActorSystem 管理所有actor实例
type ActorSystem struct {
    actors map[string]*Actor
    mu     sync.RWMutex
}

// NewActorSystem 创建一个新的actor系统
func NewActorSystem() *ActorSystem {
    return &ActorSystem{
        actors: make(map[string]*Actor),
    }
}

// RegisterActor 注册一个actor到系统
func (s *ActorSystem) RegisterActor(id string, handler MessageHandler) (ActorRef, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.actors[id]; exists {
        return nil, fmt.Errorf("actor %s already exists", id)
    }
    actor := NewActor(id, handler)
    
    s.actors[actor.id] = actor
    actor.Start()
    ref := &localActorRef{
        actor: actor,
        id:    id,
    }
    return ref, nil
}

// DeregisterActor 从系统中移除一个actor
func (s *ActorSystem) DeregisterActor(id string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if actor, exists := s.actors[id]; exists {
        actor.Stop()
        delete(s.actors, id)
    }
}

// SendMessage 发送消息到指定的actor
func (s *ActorSystem) SendMessage(to string, msg Message) error {
    s.mu.RLock()
    actor, exists := s.actors[to]
    s.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("actor %s not found", to)
    }
    
    return actor.Send(msg)
}

// Shutdown 关闭整个actor系统
func (s *ActorSystem) Shutdown() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    for _, actor := range s.actors {
        actor.Stop()
    }
    s.actors = make(map[string]*Actor)
}
