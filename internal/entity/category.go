package entity

type Category string

const (
	CategoryFood          Category = "food"
	CategoryTransport     Category = "transport"
	CategoryHousing       Category = "housing"
	CategoryEntertainment Category = "entertainment"
	CategoryHealth        Category = "health"
	CategoryShopping      Category = "shopping"
	CategoryOther         Category = "other"
)

func Categories() []Category {
	return []Category{
		CategoryFood,
		CategoryTransport,
		CategoryHousing,
		CategoryEntertainment,
		CategoryHealth,
		CategoryShopping,
		CategoryOther,
	}
}

func (c Category) Valid() bool {
	switch c {
	case CategoryFood, CategoryTransport, CategoryHousing, CategoryEntertainment,
		CategoryHealth, CategoryShopping, CategoryOther:
		return true
	default:
		return false
	}
}
