package main

type State struct {
}

func NewState() (*State, error) {
	return &State{}, nil
}
