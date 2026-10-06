package domain

// volumeDiscountTier applies to orders with at least MinQuantity stickers.
type volumeDiscountTier struct {
	MinQuantity Quantity
	BasisPoints int64
}

// Ordered from the largest tier down so the first match wins.
var volumeDiscountTiers = []volumeDiscountTier{
	{MinQuantity: 1_000, BasisPoints: 5_000},
	{MinQuantity: 500, BasisPoints: 4_000},
	{MinQuantity: 100, BasisPoints: 2_500},
	{MinQuantity: 50, BasisPoints: 1_000},
}

// Price is the breakdown shown to the customer.
type Price struct {
	Subtotal Money
	Discount Money
	Total    Money
}

// PriceFor calculates the price of a spec. It is a pure function, so the
// same spec always produces the same price and tests need no setup.
func PriceFor(spec StickerSpec) Price {
	unitPrice := pricePerSquareInch[spec.Product] * Money(spec.Size.areaSquareInches())
	subtotal := unitPrice * Money(spec.Quantity)
	discount := subtotal.percentOf(discountBasisPoints(spec.Quantity))

	return Price{
		Subtotal: subtotal,
		Discount: discount,
		Total:    subtotal - discount,
	}
}

func discountBasisPoints(quantity Quantity) int64 {
	for _, tier := range volumeDiscountTiers {
		if quantity >= tier.MinQuantity {
			return tier.BasisPoints
		}
	}
	return 0
}
