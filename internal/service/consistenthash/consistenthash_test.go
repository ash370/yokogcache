package consistenthash

import (
	"fmt"
	"strconv"
	"testing"
)

//如果要进行测试，那么我们需要明确地知道每一个传入的 key 的哈希值，
//那使用默认的 crc32.ChecksumIEEE 算法显然达不到目的。所以在这里使用了自定义的 Hash 算法。
//自定义的 Hash 算法只处理数字，传入字符串表示的数字，返回对应的数字即可。

func TestHashing(t *testing.T) {
	consistenthash := NewConsistentHash(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	consistenthash.AddTruthNodes("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if consistenthash.GetTruthNode(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	fmt.Println("23:", string(consistenthash.GetTruthNode("23")))

	// Adds 8, 18, 28
	consistenthash.AddTruthNodes("8")

	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		if consistenthash.GetTruthNode(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	fmt.Println("29:", string(consistenthash.GetTruthNode("29")))
}
