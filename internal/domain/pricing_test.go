package domain

import (
	"errors"
	"testing"
)

func TestPriceFor(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		width    int
		height   int
		quantity int
		want     Price
	}{
		{"below first tier pays full price", ProductRectangle, 3, 2, 10, Price{Subtotal: 540, Discount: 0, Total: 540}},
		{"50 units get 10%", ProductDieCut, 3, 3, 50, Price{Subtotal: 5_400, Discount: 540, Total: 4_860}},
		{"100 units get 25%", ProductCircle, 2, 2, 100, Price{Subtotal: 4_000, Discount: 1_000, Total: 3_000}},
		{"499 units stay in the 25% tier", ProductDieCut, 1, 1, 499, Price{Subtotal: 5_988, Discount: 1_497, Total: 4_491}},
		{"500 units get 40%", ProductRectangle, 1, 1, 500, Price{Subtotal: 4_500, Discount: 1_800, Total: 2_700}},
		{"1000 units get 50%", ProductCircle, 4, 4, 1_000, Price{Subtotal: 160_000, Discount: 80_000, Total: 80_000}},
		{"discount rounds down below half a cent", ProductDieCut, 1, 1, 51, Price{Subtotal: 612, Discount: 61, Total: 551}},
		{"discount rounds up at half a cent", ProductRectangle, 1, 1, 55, Price{Subtotal: 495, Discount: 50, Total: 445}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := mustSpec(t, tt.product, tt.width, tt.height, tt.quantity)

			got := PriceFor(spec)

			if got != tt.want {
				t.Errorf("PriceFor() = %+v, want %+v", got, tt.want)
			}
			if got.Total != got.Subtotal-got.Discount {
				t.Errorf("total %d is not subtotal %d minus discount %d", got.Total, got.Subtotal, got.Discount)
			}
		})
	}
}

func TestSpecValidation(t *testing.T) {
	t.Run("unknown product", func(t *testing.T) {
		if _, err := ParseProduct("holographic"); !errors.Is(err, ErrUnknownProduct) {
			t.Errorf("err = %v, want ErrUnknownProduct", err)
		}
	})

	for _, size := range [][2]int{{0, 3}, {3, 0}, {25, 1}, {1, 25}, {-1, -1}} {
		if _, err := NewSize(size[0], size[1]); !errors.Is(err, ErrInvalidSize) {
			t.Errorf("NewSize(%d, %d) err = %v, want ErrInvalidSize", size[0], size[1], err)
		}
	}
	for _, size := range [][2]int{{1, 1}, {24, 24}} {
		if _, err := NewSize(size[0], size[1]); err != nil {
			t.Errorf("NewSize(%d, %d) unexpected err = %v", size[0], size[1], err)
		}
	}

	for _, quantity := range []int{0, -5, 100_001} {
		if _, err := NewQuantity(quantity); !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("NewQuantity(%d) err = %v, want ErrInvalidQuantity", quantity, err)
		}
	}
}

func TestMoneyString(t *testing.T) {
	tests := map[Money]string{0: "$0.00", 5: "$0.05", 1_234: "$12.34", -250: "-$2.50"}
	for money, want := range tests {
		if got := money.String(); got != want {
			t.Errorf("Money(%d).String() = %q, want %q", int64(money), got, want)
		}
	}
}

func mustSpec(t *testing.T, product Product, width, height, quantity int) StickerSpec {
	t.Helper()
	size, err := NewSize(width, height)
	if err != nil {
		t.Fatal(err)
	}
	qty, err := NewQuantity(quantity)
	if err != nil {
		t.Fatal(err)
	}
	return StickerSpec{Product: product, Size: size, Quantity: qty}
}
