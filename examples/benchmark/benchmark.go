package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/actor.go"
    "time"
)

var (
    runAs       string = "all"
    listenAddr  string = ":8881"
    connectAddr string = "127.0.0.1:8881"
)

func init() {
    flag.StringVar(&runAs, "runAs", "all", "all|client|server")
    flag.StringVar(&listenAddr, "listenAddr", ":8881", "listen address")
    flag.StringVar(&connectAddr, "connectAddr", "127.0.0.1:8881", "connect address")
    flag.Parse()
}
func main() {
    fmt.Println("runAs:", runAs)
    if runAs == "client" {
        setupBenchmarkClient()
        return
    }
    if runAs == "server" {
        setupBenchmarkServer()
        select {}
        return
    }
    if runAs == "all" {
        setupBenchmarkServer()
        setupBenchmarkClient()
        return
    }
    panic("invalid runAs")
}

func setupBenchmarkServer() {
    sys := actor.NewActorSystem()
    
    _, err := sys.RegisterActor("test-actor", func(msg interface{}) (interface{}, error) {
        return "world", nil
    })
    if err != nil {
        panic(err)
    }
    
    server := actor.NewActorServer(sys, 100)
    fmt.Println("listening on", listenAddr)
    if err := server.Serve(listenAddr); err != nil {
        panic(err)
        return
    }
}

func setupBenchmarkClient() {
    ref, err := actor.NewRemoteActorRef("test-actor", connectAddr)
    if err != nil {
        panic(err)
        return
    }
    
    duration := time.Second * 30
    startTime := time.Now()
    endTime := startTime.Add(duration)
    
    var (
        successCount int64
        failCount    int64
        lastReport   = startTime
    )
    
    // 每秒报告一次进度
    go func() {
        for time.Now().Before(endTime) {
            time.Sleep(time.Second)
            current := time.Now()
            elapsed := current.Sub(lastReport)
            rate := float64(successCount) / elapsed.Seconds()
            fmt.Printf("Current rate: %.2f msg/s, Success: %d, Failed: %d\n",
                rate, successCount, failCount)
            lastReport = current
            successCount = 0 // 重置计数器
            failCount = 0
        }
    }()
    
    // 发送消息直到时间结束
    for time.Now().Before(endTime) {
        err := ref.Tell("benchmark message")
        if err != nil {
            failCount++
        } else {
            successCount++
        }
    }
    
    // 最终统计
    totalTime := time.Since(startTime)
    totalMessages := successCount + failCount
    avgRate := float64(totalMessages) / totalTime.Seconds()
    
    fmt.Printf("\nBenchmark Summary:\n")
    fmt.Printf("Total Time: %.2f seconds\n", totalTime.Seconds())
    fmt.Printf("Total Messages: %d\n", totalMessages)
    fmt.Printf("Success: %d\n", successCount)
    fmt.Printf("Failed: %d\n", failCount)
    fmt.Printf("Average Rate: %.2f msg/s\n", avgRate)
}
