//go:build cgo && !purego && integration

package zvec

import "testing"

func TestDocGetFieldsReturnErrorsForMissingAndWrongTypes(test *testing.T) {
	document := NewDoc()
	if document == nil {
		test.Fatal("NewDoc() returned nil")
	}
	defer document.Destroy()
	if err := document.AddStringField("text", "hello"); err != nil {
		test.Fatalf("AddStringField() failed: %v", err)
	}
	if err := document.AddInt64Field("number", 42); err != nil {
		test.Fatalf("AddInt64Field() failed: %v", err)
	}
	cases := []struct {
		name string
		get  func(string) error
	}{
		{"String", func(field string) error { _, err := document.GetStringField(field); return err }},
		{"Bool", func(field string) error { _, err := document.GetBoolField(field); return err }},
		{"Int32", func(field string) error { _, err := document.GetInt32Field(field); return err }},
		{"Int64", func(field string) error { _, err := document.GetInt64Field(field); return err }},
		{"Uint32", func(field string) error { _, err := document.GetUint32Field(field); return err }},
		{"Uint64", func(field string) error { _, err := document.GetUint64Field(field); return err }},
		{"Float", func(field string) error { _, err := document.GetFloatField(field); return err }},
		{"Double", func(field string) error { _, err := document.GetDoubleField(field); return err }},
		{"Vector", func(field string) error { _, err := document.GetVectorFP32Field(field); return err }},
		{"VectorInto", func(field string) error { _, err := document.GetVectorFP32FieldInto(field, nil); return err }},
	}
	for _, example := range cases {
		test.Run(example.name, func(test *testing.T) {
			if err := example.get("missing"); err == nil {
				test.Fatal("missing field returned no error")
			}
			wrongType := "number"
			if example.name == "Int64" {
				wrongType = "text"
			}
			if err := example.get(wrongType); err == nil {
				test.Fatal("wrong field type returned no error")
			}
		})
	}
	if err := document.SetFieldNull("number"); err != nil {
		test.Fatalf("SetFieldNull() failed: %v", err)
	}
	if _, err := document.GetInt64Field("number"); err == nil {
		test.Fatal("null field returned no error")
	}
}
