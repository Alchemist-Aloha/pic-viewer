package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/facette/natsort"
)

// Generate folders with realistic names to better reflect real-world performance
func generateRealisticFolders(n int) []*Folder {
	folders := make([]*Folder, n)
	prefixes := []string{"Vacation", "Trip", "Photos", "Album", "IMG", "DSC", "Project"}
	for i := 0; i < n; i++ {
		prefix := prefixes[rand.Intn(len(prefixes))]
		// Mix of upper/lower, numbers
		name := fmt.Sprintf("%s_202%d_%03d_%s", prefix, rand.Intn(5), rand.Intn(100), generateRandomString(5))
		folders[i] = &Folder{Name: name, Path: "/path/" + name}
	}
	return folders
}

var result []*Folder

// Benchmark the optimized implementation
func BenchmarkSortChildrenOptimized(b *testing.B) {
	children := generateRealisticFolders(10000)
	var r []*Folder

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		testChildren := make([]*Folder, len(children))
		copy(testChildren, children)

		sort.Slice(testChildren, func(i, j int) bool {
			return testChildren[i].Name > testChildren[j].Name
		})
		b.StartTimer()

		// THE NEW OPTIMIZED LOGIC
		type sortNode struct {
			folder *Folder
			lower  string
		}
		nodes := make([]sortNode, len(testChildren))
		for k, c := range testChildren {
			nodes[k] = sortNode{folder: c, lower: strings.ToLower(c.Name)}
		}

		sort.Slice(nodes, func(i, j int) bool {
			return natsort.Compare(nodes[i].lower, nodes[j].lower)
		})

		for k, n := range nodes {
			testChildren[k] = n.folder
		}

		r = testChildren
	}
	result = r
}

func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
