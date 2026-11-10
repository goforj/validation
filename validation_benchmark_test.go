package validation

import (
	"strings"
	"testing"
)

// BenchmarkRegistration measures a complete successful application request.
func BenchmarkRegistration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Validate(String("name", "Ada").Required().Min(2), String("email", "ada@example.com").Required().IsEmail(), Number("age", 36).Min(18), Bool("terms", true).Accepted()); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkInvalidRequest measures aggregation rather than only the first failed field.
func BenchmarkInvalidRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Validate(String("name", "").Required(), String("email", "bad").IsEmail(), Number("age", 0).Min(18)); err == nil {
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

// BenchmarkBase64Validation exposes allocations attributable to the encoded payload size.
func BenchmarkBase64Validation(b *testing.B) {
	for _, size := range []int{256, 4 << 20} {
		value := strings.Repeat("YWJj", size/4)
		chain := String("payload", value).IsBase64()
		b.Run(fmtSize(size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := Validate(chain); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// fmtSize labels the fixed resource workloads without runtime-dependent formatting.
func fmtSize(size int) string {
	if size == 256 {
		return "256B"
	}
	return "4MiB"
}

// BenchmarkSliceWithoutEach checks that parent-only validation does not scale with element count.
func BenchmarkSliceWithoutEach(b *testing.B) {
	chain := Slice("items", make([]struct{}, 1<<20)).Min(1)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := Validate(chain); err != nil {
			b.Fatal(err)
		}
	}
}
