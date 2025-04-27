// main.go
package main

import (
	"math/rand"
	"os"
	"time"
)

const tunnelListenPort = 9000 // Used on client side

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Replace with the actual IP address of the laptop/client
		runServer("192.168.88.232:9000")
	} else {
		runClient()
	}
}

