package seed

import (
	"sort"
	"time"
)

func (sd *Seed) GetBlockIndex(host Host) []uint32 {
	//TODO: make indexes distribute evenly
	ret := make([]uint32, sd.VNodeCount)
	size := uint32(len(sd.Blocks))
	for i := uint8(0); i < sd.VNodeCount; i++ {
		ret[i] = host.Hash(i) % size
	}
	// sort ret increasing order
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})

	j := 1
	n := len(ret)
	for i := 1; i < n; i++ {
		if ret[i] != ret[i-1] {
			ret[j] = ret[i]
			j++
		}
	}

	return ret[0:j]
}

func (sd *Seed) Add(host Host) {
	now := time.Now()
	sd.Lock()
	defer sd.Unlock()
	for _, h := range sd.GetBlockIndex(host) {
		block := sd.Blocks[h]
		if block.Hosts == nil {
			block.Hosts = make(map[Host]time.Time)
		}
		block.Hosts[host] = now
	}
}

func (sd *Seed) Remove(host Host) {
	sd.Lock()
	defer sd.Unlock()
	for _, h := range sd.GetBlockIndex(host) {
		block := sd.Blocks[h]
		delete(block.Hosts, host)
	}
}

func (sd *Seed) GetAllHosts() []Host {
	hosts := make([]Host, 0, sd.VNodeCount)
	sd.RLock()
	defer sd.RUnlock()
	for i := range sd.Blocks {
		for h := range sd.Blocks[i].Hosts {
			hosts = append(hosts, h)
		}
	}
	return SortDeDup(hosts)
}

//LocateBlock tries to find maximum `n` different hosts for `blockId`.
// As the golang map is not deterministic on iteration, when multiple hosts hashed
// onto the same block the LocateBlock returned hosts are also not deterministic
// which is good for load balancing.
func (sd *Seed) LocateBlock(blockId int64, n int) []Host {
	hosts := make([]Host, 0, n)
	sd.RLock()
	defer sd.RUnlock()
	for i := blockId + int64(len(sd.Blocks)); i > blockId; i-- {
		idx := i % int64(len(sd.Blocks))
		for h, _ := range sd.Blocks[idx].Hosts {
			hosts = SortDeDup(append(hosts, h))
			if len(hosts) >= n {
				return hosts
			}
		}
	}

	return hosts
}
