//go:build integration

package zvec

import "testing"

func TestDocGetVectorFP32FieldIntoReusesStorage(test *testing.T) {
	document := NewDoc()
	if document == nil {
		test.Fatal("NewDoc() returned nil")
	}
	defer document.Destroy()
	want := []float32{1, 2, 3, 4}
	if err := document.AddVectorFP32Field("embedding", want); err != nil {
		test.Fatalf("AddVectorFP32Field() failed: %v", err)
	}

	buffer := make([]float32, 0, 8)
	result, err := document.GetVectorFP32FieldInto("embedding", buffer)
	if err != nil {
		test.Fatalf("GetVectorFP32FieldInto() failed: %v", err)
	}
	if len(result) != len(want) || &result[0] != &buffer[:cap(buffer)][0] {
		test.Fatalf("GetVectorFP32FieldInto() did not reuse the destination: len=%d", len(result))
	}
	for index, value := range want {
		if result[index] != value {
			test.Fatalf("result[%d] = %v, want %v", index, result[index], value)
		}
	}

	result[0] = 99
	original, err := document.GetVectorFP32Field("embedding")
	if err != nil || original[0] != want[0] {
		test.Fatalf("returned buffer aliases the document: vector=%v, err=%v", original, err)
	}
	document.Destroy()
	if result[1] != want[1] {
		test.Fatal("returned buffer became invalid after Destroy()")
	}
}

func TestDocGetVectorFP32FieldIntoGrowthAndError(test *testing.T) {
	document := NewDoc()
	if document == nil {
		test.Fatal("NewDoc() returned nil")
	}
	defer document.Destroy()
	if err := document.AddVectorFP32Field("embedding", []float32{1, 2, 3, 4}); err != nil {
		test.Fatalf("AddVectorFP32Field() failed: %v", err)
	}

	buffer := []float32{7}
	result, err := document.GetVectorFP32FieldInto("embedding", buffer)
	if err != nil {
		test.Fatalf("GetVectorFP32FieldInto() failed: %v", err)
	}
	if len(result) != 4 || result[0] != 1 || &result[0] == &buffer[0] {
		test.Fatalf("GetVectorFP32FieldInto() did not grow the destination: %v", result)
	}
	if missing, err := document.GetVectorFP32FieldInto("missing", buffer); err == nil || missing != nil {
		test.Fatalf("missing field returned vector=%v, err=%v", missing, err)
	}
	if buffer[0] != 7 {
		test.Fatalf("failed lookup modified the destination: %v", buffer)
	}
}
