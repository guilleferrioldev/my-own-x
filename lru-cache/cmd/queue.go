package main

import "fmt"

type Queue struct {
	head   *Node
	tail   *Node
	length int
}

func NewQueue() Queue {
	head := &Node{}
	tail := &Node{}
	head.right = tail
	tail.left = head
	return Queue{head: head, tail: tail}
}

func (q *Queue) Display() {
	node := q.head.right
	if node == q.tail {
		fmt.Printf("Length %d - []\n", q.length)
		return
	}
	fmt.Printf("Length %d - [", q.length)
	for i := 0; i < q.length; i++ {
		fmt.Printf("{%s}", node.value)
		if i < q.length-1 {
			fmt.Printf("<--->")
		}
		node = node.right
	}
	fmt.Printf("]\n")
}
