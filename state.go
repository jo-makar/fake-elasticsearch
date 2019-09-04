package main

type State struct {
	pipelines map[string]string
}

func NewState() *State {
	return &State{ pipelines: make(map[string]string) }
}
