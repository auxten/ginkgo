package locater

import (
	"sync"

	"github.com/auxten/ginkgo/seed"
)

type Ring struct {
	sync.RWMutex
	Seed *seed.Seed
	//Hosts      map[seed.Host]bool
	VNodeCount uint8
}

func (r *Ring) GetBlockIndex(host seed.Host) []uint32 {
	ret := make([]uint32, r.VNodeCount)
	size := uint32(len(r.Seed.Blocks))
	for i := uint8(0); i < r.VNodeCount; i++ {
		ret[i] = host.Hash(i) % size
	}
	return ret
}

func (r *Ring) Add(host seed.Host) {
	r.Lock()
	defer r.Unlock()
	for _, h := range r.GetBlockIndex(host) {
		block := r.Seed.Blocks[h]
		if block.Hosts == nil {
			block.Hosts = make(map[seed.Host]bool)
		}
		block.Hosts[host] = true
	}
}

func (r *Ring) Remove(host seed.Host) {
	r.Lock()
	defer r.Unlock()
	for _, h := range r.GetBlockIndex(host) {
		block := r.Seed.Blocks[h]
		delete(block.Hosts, host)
	}
}

//LocateBlock tries to find maximum `n` different hosts for `blockId`.
// As the golang map is not deterministic on iteration, when multiple hosts hashed
// onto the same block the LocateBlock returned hosts are also not deterministic
// which is good for load balancing.
func (r *Ring) LocateBlock(blockId int, n int) []seed.Host {
	r.RLock()
	defer r.RUnlock()
	hosts := make([]seed.Host, 0, n)
	hostMap := make(map[seed.Host]bool) // for deduplication
blk:
	for i := blockId; i < blockId+len(r.Seed.Blocks); i++ {
		idx := i % len(r.Seed.Blocks)
		for h, _ := range r.Seed.Blocks[idx].Hosts {
			hostMap[h] = true
			if len(hostMap) >= n {
				break blk
			}
		}
	}
	for h, _ := range hostMap {
		hosts = append(hosts, h)
	}
	return hosts
}
