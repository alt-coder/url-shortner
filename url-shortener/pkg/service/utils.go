package service

import (
	"log"

	"strings"

	"github.com/go-zookeeper/zk"
)

func checkZkCounter(conn *zk.Conn) error {
	exists, _, err := conn.Exists("/counter")
	if err != nil {
		log.Printf("Error checking if /counter exists in Zookeeper: %v", err)
		return err
	}
	if !exists {
		// Using WorldACL with PermAll for simplicity. In a production environment,
		// it's highly recommended to use a more restrictive ACL with authentication.
		_, err = conn.Create("/counter", []byte("0"), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			log.Printf("Error creating /counter in Zookeeper: %v", err)
			return err
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
	for encodedBuilder.Len() < 7 {
		encodedBuilder.WriteByte('0')
	}
	return encodedBuilder.String()
}
