package main

import (
	"github.com/protolambda/ask"

	"github.com/protolambda/switcheroo/switcher"
)

func main() {
	ask.Run(&switcher.MainCmd{})
}
