package doublelist

import (
	"reflect"
	"testing"
)

func (l *List[T]) contents() []T {
	var vs []T
	iterator := l.tail
	for iterator != nil {
		v := iterator.Data
		vs = append(vs, v)
		iterator = iterator.next
	}
	return vs
}

func assertContents[T any](t *testing.T, l *List[T], want []T) {
	t.Helper()
	if !reflect.DeepEqual(l.contents(), want) {
		t.Fatalf("unexpected contents %v", l.contents())
	}
}

func TestList(t *testing.T) {
	l := &List[int]{}
	l.Append(10)
	l.Append(20)
	assertContents(t, l, []int{10, 20})
	var n *Node[int]
	n = l.Prepend(-10)
	l.Prepend(-20)
	assertContents(t, l, []int{-20, -10, 10, 20})
	l.Pop(n)
	assertContents(t, l, []int{-20, 10, 20})
	n, ok := l.PopTail()
	if !ok {
		t.Fatalf("tail should exist")
	}
	assertContents(t, l, []int{10, 20})
	if n.Data != -20 {
		t.Fatalf("unexpected data %v", n.Data)
	}
}
