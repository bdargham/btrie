package btrie

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func generateSampleHashData(size int) map[string]interface{} {
	result := make(map[string]interface{}, size)
	seed := rand.Float32()
	for i := 0; i < size; i++ {
		str := fmt.Sprintf("Here is a sample value (%d) - %f", i, seed)
		sha1 := sha1.Sum([]byte(str))
		result[string(sha1[:])] = str
	}
	return result
}

func TestBasicCrud(t *testing.T) {

	trie := New()

	data1 := generateSampleHashData(50)
	keys1 := make([]string, 0, len(data1))
    for k := range data1 {
        keys1 = append(keys1, k)
    }
	data2 := generateSampleHashData(5)

	if trie.Size() != 0 {
		t.Error("Initial trie should be empty")
	}
	for k, v := range data1 {
		if old := trie.SPut(k, v); old != nil {
			t.Errorf("Got an apparent key collision when none expected")
		}
	}

	if trie.Size() != len(data1) {
		t.Errorf("Size should be %d but instead was %d", len(data1), trie.Size())
	}

	for k, v := range data1 {
		if trie.SGet(k) != v {
			t.Error("Failed to Get value that should have been in the trie")
		}
	}

	for k, _ := range data2 {
		if trie.SGet(k) != nil {
			t.Error("Found data that shouldn't have been there")
		}
	}

	for k, _ := range data2 {
		if trie.SRemove(k) != nil {
			t.Error("Deleting inexistant entries should return nil")
		}
	}

	if trie.Size() != len(data1) {
		t.Error("Failed removals shouldn't reduce count")
	}

	idx := 0
	for _,k := range keys1[:25] {
		if ret := trie.SRemove(k); ret != data1[k] {
			t.Errorf("The value currently in trie should be returned on removal. Expected %s, got %s", data1[k], ret)
		}
		idx++
		if trie.Size() != len(data1) - idx {
			t.Error("Removing entries should reduce the size count")
		}
		if trie.SGet(k) != nil {
			t.Error("Removed entries should stay removed")
		}
	}
	// After removing half the values, check that the remaining half are still there
	for _,k := range keys1[25:] {
		if res := trie.SGet(k); res != data1[k] {
			t.Errorf("Unremoved entry should still be found in trie: %x : %s", k, res)
		}
	}

	key := "a key"
	trie.SPut(key, "Hello World")
	if old := trie.SPut(key, "Bonjour Monde"); old == nil {
		t.Error("Put with conflicting key should have returned old value")
	} else {
		if old != "Hello World" {
			t.Error("Put returned wrong old value")
		}
	}
}

func testEdgeCases(casenum int, t *testing.T, key1, key2 []byte) {
	trie := BTrie{}
	val1, val2 := "value 1", "value 2"
	trie.Put([]byte(key1), val1)
	if old := trie.Put([]byte(key2), val2); old != nil {
		t.Errorf("[%d] No old value expected", casenum)
	}
	if trie.Get([]byte(key1)) != val1 {
		t.Errorf("[%d] Error getting values. Expected [%s] but got [%s]", casenum, val1, trie.Get([]byte(key1)))
	}
	if trie.Get([]byte(key2)) != val2 {
		t.Errorf("[%d] Error getting values. Expected [%s] but got [%s]", casenum, val2, trie.Get([]byte(key2)))
	}
}

func TestEdgeCases(t *testing.T) {
	// Test case where one key is subkey of the other
	cases := [][][]byte{
		{[]byte("My-key"), []byte("My-key-longer-value")},
		{[]byte{0}, []byte{1, 1}},
		{[]byte{0}, []byte{0, 0}},
		{[]byte{0}, []byte{0, 0xFF}},
	}

	for i, pair := range cases {
		testEdgeCases(2*i, t, pair[0], pair[1])
		testEdgeCases(2*i+1, t, pair[1], pair[0])
	}

}

func TestTraversal(t *testing.T) {

	trie := BTrie{}

	sampleKey, sampleValue := "{foobar sample}", "Hello World"
	data := generateSampleHashData(50)
	data[sampleKey] = sampleValue
	for k, v := range data {
		trie.SPut(k, v)
	}

	fullTraversalResult := make([]string, 0, trie.Size())
	for cursor := trie.TraverseFully(); cursor.HasNext(); {
		entry := cursor.Next()
		entrystr := fmt.Sprintf("%x", entry.Key())
		fullTraversalResult = append(fullTraversalResult, entrystr)
	}

	if len(fullTraversalResult) != len(data) {
		t.Errorf("Traversal didn't return all entries. Expected %d and got %d", len(data), len(fullTraversalResult))
	}
	if !sort.StringsAreSorted(fullTraversalResult) {
		t.Error("Traversal didn't return entries sorted by key")
	}

	// Get index of sampleKey
	sampleKeyHex := fmt.Sprintf("%x", []byte(sampleKey))
	sampleKeyIndex := 0
	for i, v := range fullTraversalResult {
		if v == sampleKeyHex {
			sampleKeyIndex = i
			break
		}
	}

	// Do a partial Traversal
	dfsResult := make([]string, 0, trie.Size())
	for i, cursor := 0, trie.Traverse(TraversalOpts{From: []byte(sampleKey)}); cursor.HasNext(); i++ {
		entry := cursor.Next()
		if i == 0 && (!bytes.Equal(entry.Key(), []byte(sampleKey)) || entry.Value() != sampleValue) {
			t.Error("First entry returned by partial traversal doesn't correspond to search key")
		}
		dfsResult = append(dfsResult, fmt.Sprintf("%x", entry.Key()))
	}

	if !sort.StringsAreSorted(dfsResult) {
		t.Error("Partial traversal didn't return entries sorted by key")
	}

	if len(dfsResult) != len(fullTraversalResult[sampleKeyIndex:]) {
		t.Errorf("Partial traversal failed. Expected %d results and got %d", len(fullTraversalResult), len(dfsResult))
	}

	// Limited traversal
	cnt := 0
	for cursor := trie.Traverse(TraversalOpts{Limit:3}); cursor.HasNext(); {
		entry := cursor.Next()
		entrystr := fmt.Sprintf("%x", entry.Key())
		if fullTraversalResult[cnt] != entrystr {
			t.Errorf("Should return the same results as a full traversal")
		}
		cnt++
	}
	if cnt != 3 {
		t.Errorf("Traversal should respect the limit set in options")
	}

	// Backwards traversal
	cnt = 0
	for cursor := trie.Traverse(TraversalOpts{Dir:Backwards}); cursor.HasNext(); {
		entry := cursor.Next()
		entrystr := fmt.Sprintf("%x", entry.Key())
		if fullTraversalResult[len(fullTraversalResult) - 1 - cnt] != entrystr {
			t.Errorf("Backwards traversal should return in reverse order")
		}
		cnt++
	}
}
