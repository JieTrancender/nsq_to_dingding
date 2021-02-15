package main

import (
	"testing"

	"github.com/JieTrancender/nsq_to_consumer/cmd"
)

func init() {
	testing.Init()
}

func TestSystem(t *testing.T) {
	cmd.Execute()
}
