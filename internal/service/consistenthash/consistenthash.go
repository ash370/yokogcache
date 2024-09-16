package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type ConsistentHash struct {
	hash     Hash
	replicas int
	ring     []int
	peersMap map[int]string
}

func NewConsistentHash(replicas int, fn Hash) *ConsistentHash {
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}
	return &ConsistentHash{
		replicas: replicas,
		hash:     fn,
		peersMap: map[int]string{},
	}
}

func (ch *ConsistentHash) AddTruthNodes(peers ...string) {
	for _, peer := range peers {
		for i := 0; i < ch.replicas; i++ {
			//真实节点映射到虚拟节点，
			//每个真实节点创建倍数个虚拟节点
			//通过添加编号的方法区分不同虚拟节点
			hash := int(ch.hash([]byte(strconv.Itoa(i) + peer)))
			ch.ring = append(ch.ring, hash)
			ch.peersMap[hash] = peer
		}
	}
	sort.Ints(ch.ring)
}

func (ch *ConsistentHash) GetTruthNode(key string) string {
	if len(ch.ring) == 0 {
		return ""
	}

	hashVal := int(ch.hash([]byte(key)))                //key的哈希值会属于环上某个区域
	idx := sort.Search(len(ch.ring), func(i int) bool { //返回环上第一个大于key哈希值的值的下标
		return ch.ring[i] >= hashVal
	})
	return ch.peersMap[ch.ring[idx%len(ch.ring)]]
}
