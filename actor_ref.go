package actor

import (
	"context"
	"encoding/json"
	"fmt"
)

// localActorRef 表示本地actor的引用
type localActorRef struct {
	actor *Actor
	id    string
}

func (r *localActorRef) Request(ctx context.Context, req interface{}, resp interface{}) error {
	replyChan := make(chan Response, 1)
	err := r.actor.Send(Message{
		Payload: req,
		ReplyTo: replyChan,
		Context: ctx,
	})
	if err != nil {
		return err
	}

	select {
	case res := <-replyChan:
		if res.Error != "" {
			return fmt.Errorf(res.Error)
		}
		
		// 将响应数据复制到resp
		data, err := json.Marshal(res.Data)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, resp)
		
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *localActorRef) Tell(msg interface{}) error {
	return r.actor.Send(Message{
		Payload: msg,
		Context: context.Background(),
	})
}

func (r *localActorRef) ID() string {
	return r.id
}

func (r *localActorRef) Address() string {
	return "local://" + r.id
}
