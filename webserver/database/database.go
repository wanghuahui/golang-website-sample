package database

import (
	"log"

	"github.com/garyburd/redigo/redis"
)

var (
	// RedisConn var
	RedisConn redis.Conn
)

//RedisConnect func
func RedisConnect() redis.Conn {
	RedisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Println("redis Dial Error", err)
		return nil
	}
	return RedisConn
}
