package service

import (
	"log"

	"github.com/go-zookeeper/zk"
	"strings"
)

func checkZkCounter(conn *zk.Conn) error {
	exists, _, err := conn.Exists("/counter")
		if err != nil {
			log.Printf("Error checking if /counter exists in Zookeeper: %v", err)
			return err
		}
		if !exists {
			_, err = conn.Create("/counter", []byte("0"), 0, nil)
			if err != nil {
				log.Printf("Error creating /counter in Zookeeper: %v", err)
				return  err
			}
			log.Println("Created /counter in Zookeeper")
		}
		return nil
}

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func base62Encode(number int64) string {
	length := len(base62Chars)
	var encodedBuilder strings.Builder
	for number > 0 {
		remainder := number % int64(length)
		encodedBuilder.WriteByte(base62Chars[remainder])
		number /= int64(length)
	}

	return encodedBuilder.String()
}