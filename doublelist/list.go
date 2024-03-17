package doublelist

type Node[T any] struct {
	Data T
	next *Node[T]
	prev *Node[T]
}

type List[T any] struct {
	head *Node[T]
	tail *Node[T]
}

func (l *List[T]) init(v T) *Node[T] {
	l.head = &Node[T]{Data: v}
	l.tail = l.head
	return l.head
}

func (l *List[T]) Append(v T) *Node[T] {
	if l.head == nil {
		// This is the first node.
		return l.init(v)
	}
	newNode := &Node[T]{Data: v}
	newNode.prev = l.head
	l.head.next = newNode
	l.head = newNode
	return newNode
}

func (l *List[T]) Prepend(v T) *Node[T] {
	if l.tail == nil {
		return l.init(v)
	}
	newNode := &Node[T]{Data: v}
	newNode.next = l.tail
	l.tail.prev = newNode
	l.tail = newNode
	return newNode
}

func (l *List[T]) Pop(n *Node[T]) {
	if n == l.head {
		// We must move the head backwards
		l.head = n.prev
	}
	if n == l.tail {
		// We must move the tail forwards
		l.tail = n.next
	}
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}

	// Avoid any misconceptions.
	n.next, n.prev = nil, nil
}

func (l *List[T]) PopTail() (*Node[T], bool) {
	if l.tail == nil {
		return nil, false
	}
	oldTail := l.tail
	l.Pop(l.tail)
	return oldTail, true
}

func (l *List[T]) Tail() *Node[T] {
	return l.tail
}

func (l *List[T]) Head() *Node[T] {
	return l.head
}
