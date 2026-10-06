package domain

import "fmt"

// Product is the sticker cut style. Each style has its own price per square inch.
type Product string

const (
	ProductDieCut    Product = "die-cut"
	ProductCircle    Product = "circle"
	ProductRectangle Product = "rectangle"
)

var pricePerSquareInch = map[Product]Money{
	ProductDieCut:    12,
	ProductCircle:    10,
	ProductRectangle: 9,
}

// ParseProduct validates a product name coming from outside the domain.
func ParseProduct(name string) (Product, error) {
	product := Product(name)
	if _, ok := pricePerSquareInch[product]; !ok {
		return "", fmt.Errorf("%w: %q", ErrUnknownProduct, name)
	}
	return product, nil
}

const (
	minSideInches = 1
	maxSideInches = 24
	maxQuantity   = 100_000
)

// Size is a sticker size in whole inches.
type Size struct {
	WidthInches  int
	HeightInches int
}

// NewSize validates both sides against the printable range.
func NewSize(widthInches, heightInches int) (Size, error) {
	for _, side := range []int{widthInches, heightInches} {
		if side < minSideInches || side > maxSideInches {
			return Size{}, fmt.Errorf("%w: each side must be between %d and %d inches", ErrInvalidSize, minSideInches, maxSideInches)
		}
	}
	return Size{WidthInches: widthInches, HeightInches: heightInches}, nil
}

func (s Size) areaSquareInches() int64 {
	return int64(s.WidthInches) * int64(s.HeightInches)
}

// Quantity is the number of stickers in an order.
type Quantity int

// NewQuantity validates the order quantity.
func NewQuantity(value int) (Quantity, error) {
	if value < 1 || value > maxQuantity {
		return 0, fmt.Errorf("%w: must be between 1 and %d", ErrInvalidQuantity, maxQuantity)
	}
	return Quantity(value), nil
}

// StickerSpec is everything that determines the price of an order.
type StickerSpec struct {
	Product  Product
	Size     Size
	Quantity Quantity
}
