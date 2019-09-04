package main

type State struct {
	// TODO Define a fixed-size FIFO map struct for these
	pipelines, templates map[string]string
}

func NewState() *State {
	return &State{
		pipelines: make(map[string]string),
		templates: make(map[string]string),
	}
}
