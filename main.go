package main

import (
	"fmt"
	"github.com/yuanyp8/synker4harbor/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
