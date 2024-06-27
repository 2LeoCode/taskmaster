package main

import (
	"fmt"
	"taskmaster/config"
	"taskmaster/utils"
)

func main() {
	config := utils.Must(config.Parse("./tmconfig.json"))
	fmt.Printf("%+v\n", config)
}
