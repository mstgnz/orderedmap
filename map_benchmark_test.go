package orderedmap

import (
	"testing"
)

// BenchmarkSet ölçümü için
func BenchmarkSet(b *testing.B) {
	om := NewOrderedMap()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Set(i, i)
	}
}

// BenchmarkGet ölçümü için
func BenchmarkGet(b *testing.B) {
	om := NewOrderedMap()
	for i := 0; i < 1000; i++ {
		om.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Get(i % 1000)
	}
}

// BenchmarkDelete ölçümü için
func BenchmarkDelete(b *testing.B) {
	om := NewOrderedMap()
	for i := 0; i < b.N; i++ {
		om.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Delete(i)
	}
}

// BenchmarkRange ölçümü için
func BenchmarkRange(b *testing.B) {
	om := NewOrderedMap()
	for i := 0; i < 1000; i++ {
		om.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Range(func(key, value any) bool {
			return true
		})
	}
}

// BenchmarkCopy ölçümü için
func BenchmarkCopy(b *testing.B) {
	om := NewOrderedMap()
	for i := 0; i < 1000; i++ {
		om.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Copy()
	}
}

// BenchmarkParallelSet paralel set işlemleri için
func BenchmarkParallelSet(b *testing.B) {
	om := NewOrderedMap()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			om.Set(i, i)
			i++
		}
	})
}

// BenchmarkParallelGet paralel get işlemleri için
func BenchmarkParallelGet(b *testing.B) {
	om := NewOrderedMap()
	for i := 0; i < 1000; i++ {
		om.Set(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			om.Get(i % 1000)
			i++
		}
	})
}
