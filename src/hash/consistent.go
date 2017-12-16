package consistent

import (
	log "github.com/auxten/logrus"
	"hash/fnv"
	"sort"
	"github.com/Sirupsen/logrus"
)

// 一个BigBucket对应三个Bucket虚拟节点
type BigBucket struct {
	Name    string
	Bucket1 *Bucket // fnv hash 1次
	Bucket2 *Bucket // fnv hash 2次
	Bucket3 *Bucket
}

type Bucket struct {
	Position uint64
	Big      *BigBucket
}

type ConsistentRing struct {
	Range   uint64
	Buckets []Bucket
}

func (a ConsistentRing) Len() int      { return len(a.Buckets) }
func (a ConsistentRing) Swap(i, j int) { a.Buckets[i], a.Buckets[j] = a.Buckets[j], a.Buckets[i] }
func (a ConsistentRing) Less(i, j int) bool {
	return (a.Buckets[i].Position) < (a.Buckets[j].Position)
}

// 支持BigBucket
func (c *ConsistentRing) AddNode(name string) {
	h := fnv.New64()
	h.Write([]byte(name))
	c.Buckets = append(c.Buckets, Bucket{name, uint64(h.Sum64()) % c.Range})
	log.Debug(c.Buckets)
}

// 支持BigBucket
func (c ConsistentRing) DumpNodesRange() ConsistentRing {
	sort.Sort(c)
	log.Debugf("%v\n", c)
	return c
}

func (c ConsistentRing) FindBigBucketByKey(key string) (b BigBucket) {
}

func (c ConsistentRing) FindBucketByKey(key string) (b Bucket) {
	keyh := fnv.New64()
	keyh.Write([]byte(key))
	key_pos := keyh.Sum64() % c.Range
	start_bucket_idx := (sort.Search(len(c.Buckets), func(i int) bool {
		return c.Buckets[i].Position > key_pos
	}) + len(c.Buckets) - 1) % len(c.Buckets)
	b = c.Buckets[start_bucket_idx]
	start := b.Position
	end := c.Buckets[(start_bucket_idx+1)%len(c.Buckets)].Position
	logrus.Debugf("%d, start: %d, end: %d\n", key_pos, start, end)
	return
}
