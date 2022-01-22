package seed

import (
	"time"
)

func (sd *Seed) GetBlockIndex(host Host) []uint32 {
	ret := make([]uint32, sd.VNodeCount)
	size := uint32(len(sd.Blocks))
	for i := uint8(0); i < sd.VNodeCount; i++ {
		ret[i] = host.Hash(i) % size
	}
	return ret
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
func (sd *Seed) LocateBlock(blockId int, n int) []Host {
	hosts := make([]Host, 0, n)
	sd.RLock()
	defer sd.RUnlock()
	for i := blockId; i < blockId+len(sd.Blocks); i++ {
		idx := i % len(sd.Blocks)
		for h, _ := range sd.Blocks[idx].Hosts {
			hosts = SortDeDup(append(hosts, h))
			if len(hosts) >= n {
				return hosts
			}
		}
	}

	return hosts
}
