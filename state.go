package main

import (
	"fmt"
	"math/rand"
	"time"
)

type State struct {
	NodeName, ClusterName, ClusterUuid string
}

func NewState() *State {
	rand.Seed(time.Now().UnixNano())

	src := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Shuffle(len(src), func(i, j int) { src[i], src[j] = src[j], src[i] })
	uuid := string(src[:20])

	return &State{
		   NodeName: fmt.Sprintf("fake-node-%d", rand.Intn(100)),
		ClusterName: "fake-cluster",
		ClusterUuid: uuid,
	}
}
