package main

import (
	"container/heap"
	"time"
)

var queue = make(Queue, 0, 32)

type Future struct {
	ts    time.Time
	index int
	work  func()
}

type Queue []*Future

func (q Queue) Len() int { return len(q) }

func (q Queue) Less(i, j int) bool {
	return q[i].ts.Before(q[j].ts)
}

func (q Queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = j
	q[j].index = i
}

func (q *Queue) Push(v interface{}) {
	n := len(*q)
	future := v.(*Future)
	future.index = n
	*q = append(*q, future)
}

func (q *Queue) Pop() interface{} {
	old := *q
	n := len(old)
	future := old[n-1]
	future.index = -1
	*q = old[0 : n-1]
	return future
}

func At(ts time.Time, work func()) {
	heap.Push(&queue, &Future{ts: ts, work: work})
}

func After(delay time.Duration, work func()) {
	heap.Push(&queue, &Future{ts: time.Now().Add(delay), work: work})
}

func RunQueue() {
	defer log_info("Queue runner done.")
	heap.Init(&queue)
	for {
		if len(queue) == 0 {
			time.Sleep(10 * time.Microsecond)
			continue
		}
		future, ok := heap.Pop(&queue).(*Future)
		if !ok {
			log_error("there's shit on the work heap")
			continue
		}
		if future.ts.After(time.Now()) {
			time.Sleep(future.ts.Sub(time.Now()))
		}
		future.work()
	}
}
