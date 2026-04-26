package main

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/internal/cli"
)

func main() {
	flags := cli.ParseCrawlerFlags()
	fmt.Println(flags)
}
