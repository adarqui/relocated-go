package main

import (
	"log"
	"relocated_inc"
)

func main() {
	var conf relocated_inc.JsonType
	log.Printf("Initializing relocated-go.")
	conf.Init()
}
