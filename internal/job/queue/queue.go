package queue

type Queue []Executor

func NewQueue(capacity int) Queue {
	if capacity <= 0 {
		capacity = 1
	}
	pq := make(Queue, 0, capacity)
	return pq
}

func (pq Queue) Len() int {
	return len(pq)
}

func (pq *Queue) Push(item Executor) {
	n := len(*pq)
	c := cap(*pq)
	if n+1 > c {
		npq := make(Queue, n, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : n+1]
	(*pq)[n] = item
}

func (pq *Queue) Pop() Executor {
	n := len(*pq)
	c := cap(*pq)
	if n < (c/4) && c > 25 {
		npq := make(Queue, n, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[n-1]
	*pq = (*pq)[0 : n-1]
	return item
}
