package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisHost string
var redisPassword string

type RoleResp struct {
	Role   string   `json:"role"`
	Slaves []string `json:"slaves,omitempty"`
	Master string   `json:"master,omitempty"`
	Stat   string   `json:"stat,omitempty"`
}

func main() {
	flag.StringVar(&redisHost, "host", "127.0.0.1:6379", "Redis 连接地址")
	flag.StringVar(&redisPassword, "password", "", "Redis 密码")

	flag.Parse()

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := rdb.Do(ctx, "role").Result()
	if err != nil {
		panic(err)
	}

	rs := parseRoleResp(val)
	fmt.Printf("SERVER: %s\n", redisHost)
	fmt.Printf("ROLE: %s\n", rs.Role)
	switch rs.Role {
	case "master":
		fmt.Printf("SLAVES: %s\n", strings.Join(rs.Slaves, ", "))
	case "slave":
		fmt.Printf("MASTER: %s\n", rs.Master)
		fmt.Printf("STAT: %s\n", rs.Stat)
	}
}

func parseRoleResp(val interface{}) RoleResp {
	roleInfo := val.([]interface{})

	res := RoleResp{}

	role := roleInfo[0].(string)
	res.Role = role
	if role == "master" {
		res.Slaves = make([]string, 0)
		for _, ss := range roleInfo[2].([]interface{}) {
			s := ss.([]interface{})
			res.Slaves = append(res.Slaves, fmt.Sprintf("%s:%s", s[0].(string), s[1].(string)))
		}
	} else {
		res.Master = fmt.Sprintf("%s:%d", roleInfo[1].(string), roleInfo[2].(int64))
		res.Stat = roleInfo[3].(string)
	}

	return res
}
