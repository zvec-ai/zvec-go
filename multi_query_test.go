//go:build integration

package zvec

import (
	"testing"
)

func TestMultiQuerySetRerankRRF(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetRerankRRF(60); err != nil {
		t.Fatalf("SetRerankRRF failed: %v", err)
	}
}

func TestMultiQuerySetRerankWeighted(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetRerankWeighted([]float64{0.7, 0.3}); err != nil {
		t.Fatalf("SetRerankWeighted failed: %v", err)
	}
}

func TestMultiQuerySetRerankWeightedEmpty(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetRerankWeighted([]float64{}); err == nil {
		t.Error("SetRerankWeighted with empty weights should return error")
	}
}

func TestNewMultiQuery(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()
}

func TestMultiQuerySetTopK(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetTopK(10); err != nil {
		t.Fatalf("SetTopK failed: %v", err)
	}
	if got := mq.GetTopK(); got != 10 {
		t.Errorf("GetTopK() = %d, want 10", got)
	}
}

func TestMultiQuerySetFilter(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetFilter("status = 'active'"); err != nil {
		t.Fatalf("SetFilter failed: %v", err)
	}
	if got := mq.GetFilter(); got != "status = 'active'" {
		t.Errorf("GetFilter() = %q, want %q", got, "status = 'active'")
	}
}

func TestMultiQuerySetIncludeVector(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetIncludeVector(true); err != nil {
		t.Fatalf("SetIncludeVector failed: %v", err)
	}
	if !mq.GetIncludeVector() {
		t.Error("GetIncludeVector() = false, want true")
	}
}

func TestMultiQueryAddSubQuery(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	_ = sub.SetFieldName("embedding")
	_ = sub.SetNumCandidates(50)
	_ = sub.SetQueryVector([]float32{0.1, 0.2, 0.3, 0.4})

	if err := mq.AddSubQuery(sub); err != nil {
		t.Fatalf("AddSubQuery failed: %v", err)
	}
	if got := mq.GetSubQueryCount(); got != 1 {
		t.Errorf("GetSubQueryCount() = %d, want 1", got)
	}

	// sub should still be valid (copy semantics)
	if sub.handle == nil {
		t.Error("sub.handle should still be valid after AddSubQuery")
	}
}

func TestMultiQuerySetOutputFields(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	defer mq.Destroy()

	if err := mq.SetOutputFields([]string{"id", "text"}); err != nil {
		t.Fatalf("SetOutputFields failed: %v", err)
	}
}

func TestMultiQueryDestroy(t *testing.T) {
	mq := NewMultiQuery()
	if mq == nil {
		t.Fatal("NewMultiQuery returned nil")
	}
	mq.Destroy()
	mq.Destroy() // double destroy should not panic
}

func TestNewSubQuery(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()
}

func TestSubQuerySetFieldName(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	if err := sub.SetFieldName("embedding"); err != nil {
		t.Fatalf("SetFieldName failed: %v", err)
	}
	if got := sub.GetFieldName(); got != "embedding" {
		t.Errorf("GetFieldName() = %q, want %q", got, "embedding")
	}
}

func TestSubQuerySetNumCandidates(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	if err := sub.SetNumCandidates(100); err != nil {
		t.Fatalf("SetNumCandidates failed: %v", err)
	}
	if got := sub.GetNumCandidates(); got != 100 {
		t.Errorf("GetNumCandidates() = %d, want 100", got)
	}
}

func TestSubQuerySetQueryVector(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	err := sub.SetQueryVector([]float32{0.1, 0.2, 0.3, 0.4})
	if err != nil {
		t.Fatalf("SetQueryVector failed: %v", err)
	}
}

func TestSubQuerySetQueryVectorEmpty(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	err := sub.SetQueryVector([]float32{})
	if err == nil {
		t.Error("SetQueryVector with empty slice should return error")
	}
}

func TestSubQuerySetSparseVector(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	err := sub.SetSparseVector([]uint32{0, 5, 10}, []float32{0.1, 0.5, 0.9})
	if err != nil {
		t.Fatalf("SetSparseVector failed: %v", err)
	}
}

func TestSubQuerySetSparseVectorMismatch(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	err := sub.SetSparseVector([]uint32{0, 5}, []float32{0.1})
	if err == nil {
		t.Error("SetSparseVector with mismatched lengths should return error")
	}
}

func TestSubQuerySetHNSWParams(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	params := NewHNSWQueryParams(100, 0.5, false, false)
	if params == nil {
		t.Fatal("NewHNSWQueryParams returned nil")
	}

	err := sub.SetHNSWParams(params)
	if err != nil {
		t.Fatalf("SetHNSWParams failed: %v", err)
	}
	if params.handle != nil {
		t.Error("params.handle should be nil after ownership transfer")
	}
}

func TestSubQueryDestroy(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	sub.Destroy()
	sub.Destroy() // double destroy should not panic
}
