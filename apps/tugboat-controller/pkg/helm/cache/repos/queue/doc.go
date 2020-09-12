/*
Package `queue` manages a

q := queue.New()

Examples:

One
[t1-t2-t3-t4]
t1: caller invokes `q.Work(...)`
t2: queue calls Worker
t3: Worker responds
t4: caller receives error

Two callers
[t1-t2----]
[      t3-]
*/

package queue
