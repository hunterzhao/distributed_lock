package distributedlock

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

type ErrCode int

const (
	SUCCESS     ErrCode = 0
	ERR_SYSTEM  ErrCode = -1
	ERR_NOT_OWN ErrCode = -101
	ERR_IN_LOCK ErrCode = -102
)

type DistributedLock struct {
	redisPool *redis.Pool
	me        string
	timeout   int
}

type Config struct {
	redisAddress  string
	redisPassword string
	timeout       int
}

func NewDistributedLock(me string, cfg Config) DistributedLock {
	d := DistributedLock{me: me, timeout: cfg.timeout}
	d.redisPool = &redis.Pool{
		MaxIdle:     50,
		IdleTimeout: 120 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.redisAddress)
			if err != nil {
				fmt.Print(err)
				return nil, err
			}
			if _, err := c.Do("AUTH", cfg.redisPassword); err != nil {
				fmt.Print(err)
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	return d
}

func (d *DistributedLock) Lock(data1 uint64, data2 string) ErrCode {
	key := strconv.FormatUint(data1, 10) + data2
	conn := d.redisPool.Get()
	defer conn.Close()

	reply, err := redis.String(conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		fmt.Print(err)
		return ERR_SYSTEM
	}

	if reply != d.me && err != redis.ErrNil {
		return ERR_IN_LOCK
	}

	_, err = conn.Do("SETEX", key, d.timeout, d.me)
	if err != nil {
		fmt.Print(err)
		return ERR_SYSTEM
	}
	return SUCCESS
}

func (d *DistributedLock) Unlock(data1 uint64, data2 string) ErrCode {
	key := strconv.FormatUint(data1, 10) + data2
	conn := d.redisPool.Get()
	defer conn.Close()

	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		fmt.Print(err)
		if err == redis.ErrNil {
			return SUCCESS
		}
		return ERR_SYSTEM
	}

	if reply != d.me {
		return ERR_NOT_OWN
	}

	_, err = conn.Do("DEL", key)
	if err != nil {
		fmt.Print(err)
		return ERR_SYSTEM
	}

	return SUCCESS
}
