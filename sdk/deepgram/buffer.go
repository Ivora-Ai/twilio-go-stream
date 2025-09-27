package deepgram

import "fmt"

// Queue structure
type Queue struct {
	items []Item
}

type Item struct {
	data []byte
}

// Enqueue adds an element to the queue
func (q *Queue) Enqueue(item Item) {
	fmt.Println("pushed")
	q.items = append(q.items, item)
}

// Dequeue removes an element from the queue
func (q *Queue) Dequeue() (*Item, bool) {
	if len(q.items) == 0 {
		return nil, false // Queue is empty
	}
	item := q.items[0]
	q.items = q.items[1:] // Remove first element (O(n) operation)
	return &item, true
}

// IsEmpty checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}
