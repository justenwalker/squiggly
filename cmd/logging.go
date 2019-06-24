package cmd

import "log"

type logger struct {
}

func (l logger) Log(msg string) {
	log.Println(msg)
}
