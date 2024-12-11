package main

import (
    "context"
    "fmt"
    "github.com/DGHeroin/actor.go"
    "strings"
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
        err := ref.Tell("[local] hello")
        if err != nil {
            panic(err)
            return
        }
    }
    
    {
        // request
        var resp string
        err := ref.Request(context.Background(), "[rpc]do request", &resp)
        if err != nil {
            panic(err)
            return
        }
        fmt.Println("response:", resp)
    }
    
}

func setupServer() {
    sys := actor.NewActorSystem()
    
    _, err := sys.RegisterActor("test-actor", func(msg interface{}) (interface{}, error) {
        switch m := msg.(type) {
        case string, []byte:
            v := fmt.Sprintf("%v", m)
            if strings.HasPrefix(v, "[local]") {
                return "[local] world", nil
            }
            if strings.HasPrefix(v, "[rpc]") {
                return "[rpc] world", nil
            }
        }
        return "world", nil
    })
    if err != nil {
        panic(err)
    }
    
    server := actor.NewActorServer(sys, 100)
    
    if err := server.Serve("127.0.0.1:8888"); err != nil {
        panic(err)
        return
    }
}
