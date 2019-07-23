/*
Ankr Network, 2019, Song Liu

A tool to display base64 result.
*/
package main

import b64 "encoding/base64"
import "fmt"
import "flag"

func main() {
	inputPtr := flag.String("input", "", "input base64 string")
	flag.Parse()

	if *inputPtr == "" {
		fmt.Println("example: go run base64show.go -input \"MjAw\"")
	}

	sDec, _ := b64.StdEncoding.DecodeString(*inputPtr)
	fmt.Println(string(sDec))
}
