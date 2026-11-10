package validation

import "testing"

// BenchmarkRegistration measures a complete successful application request.
func BenchmarkRegistration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Validate(String("name", "Ada").Required().Min(2), String("email", "ada@example.com").Required().Email(), Number("age", 36).Min(18), Bool("terms", true).Accepted()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvalidRequest measures aggregation rather than only the first failed field.
func BenchmarkInvalidRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Validate(String("name", "").Required(), String("email", "bad").Email(), Number("age", 0).Min(18)); err == nil {
			b.Fatal("invalid input passed")
		}
	}
}

// BenchmarkNestedItems measures a request with one hundred independently validated items.
func BenchmarkNestedItems(b *testing.B) {
	values := make([]int, 100)
	for i := range values {
		values[i] = i + 1
	}
	for i := 0; i < b.N; i++ {
		if err := Validate(Slice("items", values).Required().Each(func(v int) []Field { return []Field{Number("quantity", v).Min(1)} })); err != nil {
			b.Fatal(err)
		}
	}
}
