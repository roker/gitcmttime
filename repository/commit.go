package repository

import (
	"github.com/go-git/go-git/v5/plumbing"
	"time"
)

type commitStruct struct {
	id       plumbing.Hash
	when     time.Time
	children map[plumbing.Hash]*commitStruct
	parents  map[plumbing.Hash]*commitStruct
}

type changeStruct struct {
	*commitStruct
	new  time.Time
	from plumbing.Hash
}

// adds a new child object if it doesnt exist yet or is not null
func (c *commitStruct) addChild(cs *commitStruct) {
	if cs == nil {
		return
	}

	if _, exists := c.children[cs.id]; exists {
		return
	}

	c.children[cs.id] = cs
}
