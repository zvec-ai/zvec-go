//go:build integration

package zvec

import (
	"testing"
)

func TestNewHNSWQueryParams(t *testing.T) {
	params := NewHNSWQueryParams(100, 0.5, false, false)
	if params == nil {
		t.Fatal("NewHNSWQueryParams returned nil")
	}
	if params.GetEf() != 100 {
		t.Errorf("Expected GetEf() to return 100, got %d", params.GetEf())
	}
	params.Destroy()
}

func TestHNSWQueryParamsSetEf(t *testing.T) {
	params := NewHNSWQueryParams(100, 0.5, false, false)
	if params == nil {
		t.Fatal("NewHNSWQueryParams returned nil")
	}
	defer params.Destroy()

	err := params.SetEf(200)
	if err != nil {
		t.Errorf("SetEf failed: %v", err)
	}
	if params.GetEf() != 200 {
		t.Errorf("Expected GetEf() to return 200 after SetEf, got %d", params.GetEf())
	}
}

func TestHNSWQueryParamsDestroy(t *testing.T) {
	params := NewHNSWQueryParams(100, 0.5, false, false)
	if params == nil {
		t.Fatal("NewHNSWQueryParams returned nil")
	}

	// First Destroy should not panic
	params.Destroy()

	// Second Destroy should not panic
	params.Destroy()
}

func TestNewIVFQueryParams(t *testing.T) {
	params := NewIVFQueryParams(10, false, 1.0)
	if params == nil {
		t.Fatal("NewIVFQueryParams returned nil")
	}
	params.Destroy()
}

func TestIVFQueryParamsSetNprobe(t *testing.T) {
	params := NewIVFQueryParams(10, false, 1.0)
	if params == nil {
		t.Fatal("NewIVFQueryParams returned nil")
	}
	defer params.Destroy()

	err := params.SetNprobe(20)
	if err != nil {
		t.Errorf("SetNprobe failed: %v", err)
	}
}

func TestIVFQueryParamsDestroy(t *testing.T) {
	params := NewIVFQueryParams(10, false, 1.0)
	if params == nil {
		t.Fatal("NewIVFQueryParams returned nil")
	}

	// First Destroy should not panic
	params.Destroy()

	// Second Destroy should not panic
	params.Destroy()
}

func TestNewFlatQueryParams(t *testing.T) {
	params := NewFlatQueryParams(false, 1.0)
	if params == nil {
		t.Fatal("NewFlatQueryParams returned nil")
	}
	params.Destroy()
}

func TestFlatQueryParamsDestroy(t *testing.T) {
	params := NewFlatQueryParams(false, 1.0)
	if params == nil {
		t.Fatal("NewFlatQueryParams returned nil")
	}

	// First Destroy should not panic
	params.Destroy()

	// Second Destroy should not panic
	params.Destroy()
}

func TestNewSearchQuery(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	query.Destroy()
}

func TestSearchQueryFieldName(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetFieldName("vector_field")
	if err != nil {
		t.Errorf("SetFieldName failed: %v", err)
	}

	if query.GetFieldName() != "vector_field" {
		t.Errorf("Expected GetFieldName() to return 'vector_field', got '%s'", query.GetFieldName())
	}
}

func TestSearchQueryTopK(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetTopK(10)
	if err != nil {
		t.Errorf("SetTopK failed: %v", err)
	}

	if query.GetTopK() != 10 {
		t.Errorf("Expected GetTopK() to return 10, got %d", query.GetTopK())
	}
}

func TestSearchQuerySetQueryVector(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	vector := []float32{1.0, 2.0, 3.0, 4.0}
	err := query.SetQueryVector(vector)
	if err != nil {
		t.Errorf("SetQueryVector failed: %v", err)
	}
}

func TestSearchQuerySetQueryVectorEmpty(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	vector := []float32{}
	err := query.SetQueryVector(vector)
	if err == nil {
		t.Error("Expected error when setting empty vector, got nil")
	}
}

func TestSearchQueryFilter(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	filter := "category == 'test'"
	err := query.SetFilter(filter)
	if err != nil {
		t.Errorf("SetFilter failed: %v", err)
	}

	if query.GetFilter() != filter {
		t.Errorf("Expected GetFilter() to return '%s', got '%s'", filter, query.GetFilter())
	}
}

func TestSearchQueryIncludeVector(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetIncludeVector(true)
	if err != nil {
		t.Errorf("SetIncludeVector failed: %v", err)
	}

	if !query.GetIncludeVector() {
		t.Error("Expected GetIncludeVector() to return true, got false")
	}
}

func TestSearchQueryIncludeDocID(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetIncludeDocID(true)
	if err != nil {
		t.Errorf("SetIncludeDocID failed: %v", err)
	}

	if !query.GetIncludeDocID() {
		t.Error("Expected GetIncludeDocID() to return true, got false")
	}
}

func TestSearchQueryOutputFields(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	fields := []string{"field1", "field2", "field3"}
	err := query.SetOutputFields(fields)
	if err != nil {
		t.Errorf("SetOutputFields failed: %v", err)
	}
}

func TestSearchQuerySetHNSWParams(t *testing.T) {
	params := NewHNSWQueryParams(100, 0.5, false, false)
	if params == nil {
		t.Fatal("NewHNSWQueryParams returned nil")
	}

	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetHNSWParams(params)
	if err != nil {
		t.Errorf("SetHNSWParams failed: %v", err)
	}

	// Verify ownership transfer: params.handle should be nil
	if params.handle != nil {
		t.Error("Expected params.handle to be nil after ownership transfer, got non-nil")
	}
}

func TestSearchQuerySetIVFParams(t *testing.T) {
	params := NewIVFQueryParams(10, false, 1.0)
	if params == nil {
		t.Fatal("NewIVFQueryParams returned nil")
	}

	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetIVFParams(params)
	if err != nil {
		t.Errorf("SetIVFParams failed: %v", err)
	}

	// Verify ownership transfer: params.handle should be nil
	if params.handle != nil {
		t.Error("Expected params.handle to be nil after ownership transfer, got non-nil")
	}
}

func TestSearchQuerySetFlatParams(t *testing.T) {
	params := NewFlatQueryParams(false, 1.0)
	if params == nil {
		t.Fatal("NewFlatQueryParams returned nil")
	}

	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetFlatParams(params)
	if err != nil {
		t.Errorf("SetFlatParams failed: %v", err)
	}

	// Verify ownership transfer: params.handle should be nil
	if params.handle != nil {
		t.Error("Expected params.handle to be nil after ownership transfer, got non-nil")
	}
}

func TestSearchQueryDestroy(t *testing.T) {
	query := NewSearchQuery()
	if query == nil {
		t.Fatal("NewSearchQuery returned nil")
	}

	// First Destroy should not panic
	query.Destroy()

	// Second Destroy should not panic
	query.Destroy()
}

func TestNewGroupBySearchQuery(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	query.Destroy()
}

func TestGroupBySearchQuerySetFieldName(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetFieldName("vector_field")
	if err != nil {
		t.Errorf("SetFieldName failed: %v", err)
	}
}

func TestGroupBySearchQuerySetGroupByFieldName(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetGroupByFieldName("group_field")
	if err != nil {
		t.Errorf("SetGroupByFieldName failed: %v", err)
	}
}

func TestGroupBySearchQuerySetGroupCount(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetGroupCount(10)
	if err != nil {
		t.Errorf("SetGroupCount failed: %v", err)
	}
}

func TestGroupBySearchQuerySetTopkPerGroup(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetTopkPerGroup(5)
	if err != nil {
		t.Errorf("SetTopkPerGroup failed: %v", err)
	}
}

func TestGroupBySearchQuerySetQueryVector(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	vector := []float32{1.0, 2.0, 3.0, 4.0}
	err := query.SetQueryVector(vector)
	if err != nil {
		t.Errorf("SetQueryVector failed: %v", err)
	}
}

func TestGroupBySearchQuerySetQueryVectorEmpty(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	vector := []float32{}
	err := query.SetQueryVector(vector)
	if err == nil {
		t.Error("Expected error when setting empty vector, got nil")
	}
}

func TestGroupBySearchQuerySetFilter(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	filter := "category == 'test'"
	err := query.SetFilter(filter)
	if err != nil {
		t.Errorf("SetFilter failed: %v", err)
	}
}

func TestGroupBySearchQuerySetIncludeVector(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetIncludeVector(true)
	if err != nil {
		t.Errorf("SetIncludeVector failed: %v", err)
	}
}

func TestGroupBySearchQuerySetOutputFields(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	fields := []string{"field1", "field2", "field3"}
	err := query.SetOutputFields(fields)
	if err != nil {
		t.Errorf("SetOutputFields failed: %v", err)
	}
}

func TestGroupBySearchQuerySetHNSWParams(t *testing.T) {
	params := NewHNSWQueryParams(100, 0.5, false, false)
	if params == nil {
		t.Fatal("NewHNSWQueryParams returned nil")
	}

	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetHNSWParams(params)
	if err != nil {
		t.Errorf("SetHNSWParams failed: %v", err)
	}

	// Verify ownership transfer: params.handle should be nil
	if params.handle != nil {
		t.Error("Expected params.handle to be nil after ownership transfer, got non-nil")
	}
}

func TestGroupBySearchQuerySetIVFParams(t *testing.T) {
	params := NewIVFQueryParams(10, false, 1.0)
	if params == nil {
		t.Fatal("NewIVFQueryParams returned nil")
	}

	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetIVFParams(params)
	if err != nil {
		t.Errorf("SetIVFParams failed: %v", err)
	}

	// Verify ownership transfer: params.handle should be nil
	if params.handle != nil {
		t.Error("Expected params.handle to be nil after ownership transfer, got non-nil")
	}
}

func TestGroupBySearchQuerySetFlatParams(t *testing.T) {
	params := NewFlatQueryParams(false, 1.0)
	if params == nil {
		t.Fatal("NewFlatQueryParams returned nil")
	}

	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}
	defer query.Destroy()

	err := query.SetFlatParams(params)
	if err != nil {
		t.Errorf("SetFlatParams failed: %v", err)
	}

	// Verify ownership transfer: params.handle should be nil
	if params.handle != nil {
		t.Error("Expected params.handle to be nil after ownership transfer, got non-nil")
	}
}

func TestGroupBySearchQueryDestroy(t *testing.T) {
	query := NewGroupBySearchQuery()
	if query == nil {
		t.Fatal("NewGroupBySearchQuery returned nil")
	}

	// First Destroy should not panic
	query.Destroy()

	// Second Destroy should not panic
	query.Destroy()
}
