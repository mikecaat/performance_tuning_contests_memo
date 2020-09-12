package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

// コネクションプーリング
func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	pool *redis.Pool
)

func main() {
	pool = newPool("redis:6379")

	router := gin.Default()

	// String型 普通のkey-value
	router.GET("/string", func(context *gin.Context) {

		conn := pool.Get()
		defer conn.Close()

		// 1件づつ同期的に格納
		// The Do method combines the functionality of the Send, Flush and Receive methods.
		conn.Do("SET", "Hello", "World")

		// データの取得
		v, err := conn.Do("GET", "Hello")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("get: %s\n", v)

		// 非同期(pipeline)でまとめて格納. Send/Flush/Receiveをそれぞれ呼ぶ
		for i := 0; i < 100000; i++ {
			conn.Send("SET", i, i)
		}
		conn.Flush() //まとめて格納
		conn.Receive()

		context.JSON(200, gin.H{
			"message": "hello world",
		})
	})

	// set型 特定の値が存在するかのチェックに使える
	router.GET("/set", func(context *gin.Context) {

		conn := pool.Get()
		defer conn.Close()

		// pipelineでデータ格納
		set := "SET_TYPE"
		for i := 0; i < 100000; i++ {
			conn.Send("SADD", set, i)
		}
		conn.Flush()
		conn.Receive()

		// 存在チェック
		b, _ := conn.Do("SISMEMBER", set, 100)
		fmt.Printf("100 is exists? : %d\n", b)

		// 試しに消してから存在チェック
		conn.Do("SREM", set, 100)
		fmt.Printf("100 removed\n")

		b, _ = conn.Do("SISMEMBER", set, 100)
		fmt.Printf("100 is exists? : %d\n", b)

		context.JSON(200, gin.H{
			"message": "ok",
		})
	})

	router.GET("/flushall", func(context *gin.Context) {

		conn := pool.Get()
		defer conn.Close()

		// データの全削除
		conn.Do("FLUSHALL")

		context.JSON(200, gin.H{
			"message": "ok",
		})
	})

	router.Run(":8080")
}
