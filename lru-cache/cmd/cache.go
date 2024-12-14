package main

import "fmt"

const (
	SIZE = 10
)

type Node struct {
	value string
	left  *Node
	right *Node
}

type Hash map[string]*Node

type Cache struct {
	queue Queue
	hash  Hash
}

func NewCache() *Cache {
	return &Cache{queue: NewQueue(), hash: Hash{}}
}

func (c *Cache) Check(str string) {
	if val, ok := c.hash[str]; ok {
		c.Remove(val)
	}
	node := &Node{value: str}
	c.Add(node)
	c.hash[str] = node
}

func (c *Cache) Add(node *Node) {
	fmt.Printf("Adding %s\n", node.value)
	temp := c.queue.head.right
	c.queue.head.right = node
	node.left = c.queue.head
	node.right = temp
	temp.left = node
	c.queue.length++

	if c.queue.length > SIZE {
		c.Remove(c.queue.tail.left)
	}
}

func (c *Cache) Remove(node *Node) {
	fmt.Printf("Removing %s\n", node.value)
	if node == nil {
		return
	}
	left := node.left
	right := node.right
	left.right = right
	right.left = left
	c.queue.length--
	delete(c.hash, node.value)
}

func (c *Cache) Display() {
	fmt.Printf("Cache:\n")
	c.queue.Display()
}
