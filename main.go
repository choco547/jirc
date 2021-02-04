package main

import (
	"math/rand"
	"time"

	"github.com/ikeberlein/jirc/cmd"
)

func main() {
	rand.Seed(time.Now().Unix())
	cmd.Execute()
}
