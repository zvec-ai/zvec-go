//go:build integration

package zvec

import (
	"testing"
)

// =============================================================================
// DiskANN IndexParams Tests
// =============================================================================

func TestNewDiskANNIndexParams(t *testing.T) {
	params, err := NewDiskANNIndexParams(MetricTypeL2, 100, 50, 0)
	if err != nil {
		t.Fatalf("NewDiskANNIndexParams() failed: %v", err)
	}
	defer params.Destroy()

	if got := params.GetType(); got != IndexTypeDiskANN {
		t.Errorf("GetType() = %v, want %v", got, IndexTypeDiskANN)
	}

	if got := params.GetMetricType(); got != MetricTypeL2 {
		t.Errorf("GetMetricType() = %v, want %v", got, MetricTypeL2)
	}

	if got := params.GetDiskANNMaxDegree(); got != 100 {
		t.Errorf("GetDiskANNMaxDegree() = %d, want %d", got, 100)
	}

	if got := params.GetDiskANNListSize(); got != 50 {
		t.Errorf("GetDiskANNListSize() = %d, want %d", got, 50)
	}

	if got := params.GetDiskANNPQChunkNum(); got != 0 {
		t.Errorf("GetDiskANNPQChunkNum() = %d, want %d", got, 0)
	}
}

func TestDiskANNIndexParamsSetters(t *testing.T) {
	params, err := NewDiskANNIndexParams(MetricTypeCosine, 64, 32, 0)
	if err != nil {
		t.Fatalf("NewDiskANNIndexParams() failed: %v", err)
	}
	defer params.Destroy()

	// Test SetDiskANNParams round-trip
	if err := params.SetDiskANNParams(128, 64, 16); err != nil {
		t.Fatalf("SetDiskANNParams() failed: %v", err)
	}
	if got := params.GetDiskANNMaxDegree(); got != 128 {
		t.Errorf("GetDiskANNMaxDegree() = %d, want %d", got, 128)
	}
	if got := params.GetDiskANNListSize(); got != 64 {
		t.Errorf("GetDiskANNListSize() = %d, want %d", got, 64)
	}
	if got := params.GetDiskANNPQChunkNum(); got != 16 {
		t.Errorf("GetDiskANNPQChunkNum() = %d, want %d", got, 16)
	}
}

func TestDiskANNIndexParamsDestroy(t *testing.T) {
	params, err := NewDiskANNIndexParams(MetricTypeL2, 100, 50, 0)
	if err != nil {
		t.Fatalf("NewDiskANNIndexParams() failed: %v", err)
	}

	// First Destroy should not panic
	params.Destroy()

	// Second Destroy should also not panic
	params.Destroy()
}

func TestDiskANNIndexParamsAllMetricTypes(t *testing.T) {
	testTypes := []MetricType{
		MetricTypeL2,
		MetricTypeIP,
		MetricTypeCosine,
	}
	for _, metric := range testTypes {
		params, err := NewDiskANNIndexParams(metric, 100, 50, 0)
		if err != nil {
			t.Errorf("NewDiskANNIndexParams(%v) failed: %v", metric, err)
			continue
		}
		if got := params.GetMetricType(); got != metric {
			t.Errorf("GetMetricType() = %v, want %v", got, metric)
		}
		params.Destroy()
	}
}

// =============================================================================
// DiskANNQueryParams Tests
// =============================================================================

func TestNewDiskANNQueryParams(t *testing.T) {
	params := NewDiskANNQueryParams(300)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	if got := params.GetListSize(); got != 300 {
		t.Errorf("GetListSize() = %d, want %d", got, 300)
	}
	params.Destroy()
}

func TestDiskANNQueryParamsSetListSize(t *testing.T) {
	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	defer params.Destroy()

	if err := params.SetListSize(200); err != nil {
		t.Errorf("SetListSize failed: %v", err)
	}
	if got := params.GetListSize(); got != 200 {
		t.Errorf("GetListSize() = %d, want %d", got, 200)
	}
}

func TestDiskANNQueryParamsSetRadius(t *testing.T) {
	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	defer params.Destroy()

	if err := params.SetRadius(0.5); err != nil {
		t.Errorf("SetRadius failed: %v", err)
	}
	if got := params.GetRadius(); got != 0.5 {
		t.Errorf("GetRadius() = %f, want %f", got, 0.5)
	}
}

func TestDiskANNQueryParamsSetIsLinear(t *testing.T) {
	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	defer params.Destroy()

	if err := params.SetIsLinear(true); err != nil {
		t.Errorf("SetIsLinear failed: %v", err)
	}
	if got := params.GetIsLinear(); got != true {
		t.Errorf("GetIsLinear() = %v, want %v", got, true)
	}
}

func TestDiskANNQueryParamsSetIsUsingRefiner(t *testing.T) {
	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	defer params.Destroy()

	if err := params.SetIsUsingRefiner(true); err != nil {
		t.Errorf("SetIsUsingRefiner failed: %v", err)
	}
	if got := params.GetIsUsingRefiner(); got != true {
		t.Errorf("GetIsUsingRefiner() = %v, want %v", got, true)
	}
}

func TestDiskANNQueryParamsDestroy(t *testing.T) {
	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}

	// First Destroy should not panic
	params.Destroy()

	// Second Destroy should not panic
	params.Destroy()
}

// =============================================================================
// SetDiskANNParams ownership transfer Tests
// =============================================================================

func TestSearchQuerySetDiskANNParams(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	if err := query.SetFieldName("vector_field"); err != nil {
		t.Fatalf("SetFieldName failed: %v", err)
	}

	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	if err := query.SetDiskANNParams(params); err != nil {
		t.Fatalf("SetDiskANNParams failed: %v", err)
	}
	if params.handle != nil {
		t.Fatal("SetDiskANNParams did not transfer ownership")
	}

	// Destroy after ownership transfer should not panic
	params.Destroy()
}

func TestGroupBySearchQuerySetDiskANNParams(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	if err := query.SetFieldName("vector_field"); err != nil {
		t.Fatalf("SetFieldName failed: %v", err)
	}

	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	if err := query.SetDiskANNParams(params); err != nil {
		t.Fatalf("SetDiskANNParams failed: %v", err)
	}
	if params.handle != nil {
		t.Fatal("SetDiskANNParams did not transfer ownership")
	}

	params.Destroy()
}

func TestSubQuerySetDiskANNParams(t *testing.T) {
	sub := NewSubQuery()
	if sub == nil {
		t.Fatal("NewSubQuery returned nil")
	}
	defer sub.Destroy()

	if err := sub.SetFieldName("vector_field"); err != nil {
		t.Fatalf("SetFieldName failed: %v", err)
	}

	params := NewDiskANNQueryParams(100)
	if params == nil {
		t.Fatal("NewDiskANNQueryParams returned nil")
	}
	if err := sub.SetDiskANNParams(params); err != nil {
		t.Fatalf("SetDiskANNParams failed: %v", err)
	}
	if params.handle != nil {
		t.Fatal("SetDiskANNParams did not transfer ownership")
	}

	params.Destroy()
}

func TestSetDiskANNParamsNilParam(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	// nil params should return error
	if err := query.SetDiskANNParams(nil); err == nil {
		t.Fatal("SetDiskANNParams(nil) should return error")
	}
}
