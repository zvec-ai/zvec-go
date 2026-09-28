//go:build integration

package zvec

import (
	"fmt"
	"testing"
)

func BenchmarkDocGetVectorFP32FieldComparison(benchmark *testing.B) {
	for _, dimension := range []int{128, 768} {
		benchmark.Run(fmt.Sprintf("Dim%d", dimension), func(dimensionBenchmark *testing.B) {
			document := NewDoc()
			if document == nil {
				dimensionBenchmark.Fatal("NewDoc() returned nil")
			}
			defer document.Destroy()
			if err := document.AddVectorFP32Field("embedding", generateRandomVector(dimension)); err != nil {
				dimensionBenchmark.Fatalf("AddVectorFP32Field() failed: %v", err)
			}
			dimensionBenchmark.Run("Allocating", func(getBenchmark *testing.B) {
				getBenchmark.ReportAllocs()
				getBenchmark.ResetTimer()
				for iteration := 0; iteration < getBenchmark.N; iteration++ {
					vector, err := document.GetVectorFP32Field("embedding")
					if err != nil || len(vector) != dimension {
						getBenchmark.Fatalf("GetVectorFP32Field() returned %d dimensions, err=%v", len(vector), err)
					}
				}
			})
			dimensionBenchmark.Run("Into", func(getBenchmark *testing.B) {
				buffer := make([]float32, dimension)
				getBenchmark.ReportAllocs()
				getBenchmark.ResetTimer()
				for iteration := 0; iteration < getBenchmark.N; iteration++ {
					vector, err := document.GetVectorFP32FieldInto("embedding", buffer)
					if err != nil || len(vector) != dimension {
						getBenchmark.Fatalf("GetVectorFP32FieldInto() returned %d dimensions, err=%v", len(vector), err)
					}
				}
			})
		})
	}
}

func BenchmarkDocGetFixedFields(benchmark *testing.B) {
	document := NewDoc()
	if document == nil {
		benchmark.Fatal("NewDoc() returned nil")
	}
	benchmark.Cleanup(document.Destroy)
	if err := document.AddStringField("content", "benchmark content"); err != nil {
		benchmark.Fatalf("AddStringField() failed: %v", err)
	}
	if err := document.AddInt64Field("count", 42); err != nil {
		benchmark.Fatalf("AddInt64Field() failed: %v", err)
	}
	benchmark.Run("String", func(fieldBenchmark *testing.B) {
		fieldBenchmark.ReportAllocs()
		for iteration := 0; iteration < fieldBenchmark.N; iteration++ {
			value, err := document.GetStringField("content")
			if err != nil || value != "benchmark content" {
				fieldBenchmark.Fatalf("GetStringField() returned %q, err=%v", value, err)
			}
		}
	})
	benchmark.Run("Int64", func(fieldBenchmark *testing.B) {
		fieldBenchmark.ReportAllocs()
		for iteration := 0; iteration < fieldBenchmark.N; iteration++ {
			value, err := document.GetInt64Field("count")
			if err != nil || value != 42 {
				fieldBenchmark.Fatalf("GetInt64Field() returned %d, err=%v", value, err)
			}
		}
	})
}
