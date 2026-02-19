package cache

type Shop struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CurrencyCode string `json:"currency_code"`
}

func BuildShop(
	id string,
	name string,
	currencyCode string,
) *Shop {
	return &Shop{
		ID:           id,
		Name:         name,
		CurrencyCode: currencyCode,
	}
}
