package main

import (
    "context"
    "fmt"
    "github.com/DGHeroin/actor.go"
    "time"
)

func main() {
    setupServer()
    setupClient()
    for {
        time.Sleep(time.Second)
    }
}

func setupClient() {
    ref, err := actor.NewRemoteActorRef("test-actor", "127.0.0.1:8888")
    if err != nil {
        panic(err)
        return
    }
    {
        // tell
        err := ref.Tell("say hello")
        if err != nil {
            panic(err)
            return
        }
    }
    
    {
        // request
        var resp string
        err := ref.Request(context.Background(), "do request", &resp)
        if err != nil {
            panic(err)
            return
        }
        fmt.Println("response:", resp)
    }
    
}

func setupServer() {
    sys := actor.NewActorSystem()
    
    ref, err := sys.RegisterActor("test-actor", func(msg interface{}) (interface{}, error) {
        switch m := msg.(type) {
        case string, []byte:
            fmt.Printf("received message:%s\n", m)
        }
        return "world", nil
    })
    if err != nil {
        panic(err)
    }
    var resp string
    if err := ref.Request(context.Background(), "hello", &resp); err != nil {
        panic(err)
        return
    }
    fmt.Println("response:", resp)
    server := actor.NewActorServer(sys)
    
    if err := server.Serve("127.0.0.1:8888"); err != nil {
        panic(err)
        return
    }
}
