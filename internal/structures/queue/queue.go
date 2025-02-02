package queue

type queueNode[T any] struct {
	data T
	next *queueNode[T]
}

type Queue[T any] struct {
	head  *queueNode[T]
	tail  *queueNode[T]
	count int
}

func New[T any]() *Queue[T] {
	return &Queue[T]{
		head:  nil,
		tail:  nil,
		count: 0,
	}
}

func (q *Queue[T]) Enqueue(data T) *Queue[T] {
	node := &queueNode[T]{
		data: data,
		next: nil,
	}

	if q.head == nil {
		q.head = node
		q.tail = node
	} else {
		q.tail.next = node
		q.tail = node
	}

	q.count++
	return q
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if q.head == nil {
		return *new(T), false
	}

	data := q.head.data
	q.head = q.head.next
	q.count--
	return data, true
}
