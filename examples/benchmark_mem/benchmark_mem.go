package main

import (
    "fmt"
    "time"
    
    "github.com/DGHeroin/actor.go"
)

var actorRef actor.ActorRef

func main() {
    setupBenchmarkServer()
    setupBenchmarkClient()
}

func setupBenchmarkServer() {
    sys := actor.NewActorSystem()
    
    ref, err := sys.RegisterActor("test-actor", func(msg interface{}) (interface{}, error) {
        return "world", nil
    })
    if err != nil {
        panic(err)
    }
    actorRef = ref
}

func setupBenchmarkClient() {
    ref := actorRef
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
        lastSuccessCount := successCount
        lastFailCount := failCount
        for time.Now().Before(endTime) {
            time.Sleep(time.Second)
            current := time.Now()
            elapsed := current.Sub(lastReport)
            
            successRate := float64(successCount-lastSuccessCount) / elapsed.Seconds()
            failRate := float64(failCount-lastFailCount) / elapsed.Seconds()
            
            lastReport = current
            lastSuccessCount = successCount
            lastFailCount = failCount
            
            fmt.Printf("\r%.2f msg/s, %.2f msg/s failed", successRate, failRate)
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
