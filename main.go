package main

import (
	"taskmaster/parsing"
	"taskmaster/utils"
)

func main() {
	config := utils.Must(parsing.ParseConfig("./tmconfig.json"))
	println("Hello world!")
}
