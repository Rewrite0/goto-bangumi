package parser

import (
	"bytes"
	_ "embed"
	"testing"

	"golang.org/x/net/html"
)

//go:embed testdata/mikan_3599.html
var benchMikan3599HTML []byte

//go:embed testdata/mikan_3751.html
var benchMikan3751HTML []byte

//go:embed testdata/mikan_3790.html
var benchMikan3790HTML []byte

func BenchmarkFindRSSLink(b *testing.B) {
	parser := NewMikanParser()

	benchmarks := []struct {
		name     string
		htmlData []byte
	}{
		{"mikan_3599", benchMikan3599HTML},
		{"mikan_3751", benchMikan3751HTML},
		{"mikan_3790", benchMikan3790HTML},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				doc, _ := html.Parse(bytes.NewReader(bm.htmlData))
				_ = parser.findRSSLink(doc)
			}
		})
	}
}

// BenchmarkFindRSSLink_Parallel 测试并发性能
func BenchmarkFindRSSLink_Parallel(b *testing.B) {
	parser := NewMikanParser()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doc, _ := html.Parse(bytes.NewReader(benchMikan3751HTML))
			_ = parser.findRSSLink(doc)
		}
	})
}
