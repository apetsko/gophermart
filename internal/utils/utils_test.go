package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func random(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		log.Println(err)
	}
	return nBig.Int64()
}

func generateURLS(count int) []string {
	urls := make([]string, count)
	protocols := []string{
		"http://",
		"https://",
		//"ftp://",
		//"ftps://",
		//"sftp://",
		//"mailto://",
		//"telnet://",
		//"file://",
		//"data://",
		//"ws://",
		//"wss://",
		//"bluetooth://",
	}

	for i := range urls {
		u := fmt.Sprintf("%s%s.%s/%s/%s",
			protocols[random(int64(len(protocols)))], //protocol
			randomString(random(int64(10))+1),        //2domain
			randomString(random(int64(3))+2),         //1domain
			randomString(random(int64(9))+2),         //path1
			randomString(random(int64(13))+4),        //path2
		)
		urls[i] = u
	}

	return urls
}

func randomString(length int64) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[random(int64(len(charset)))]
	}
	return string(b)
}

func Test_Generate(t *testing.T) {
	m := make(map[string]string)
	urls := generateURLS(100000)
	for i, u := range urls {
		t.Run(fmt.Sprintf("URL #%d", i), func(t *testing.T) {
			IDlen := 8
			ID := GenerateID(u, IDlen)

			ur, ok := m[ID]
			require.Equal(t, false, ok)
			require.Equal(t, false, ur == u)

			m[ID] = u
		})
	}
}
