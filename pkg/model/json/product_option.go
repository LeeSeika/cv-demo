package json

type ProductOption struct {
	ID string `json:"id"`
	// The product option's name.
	Name string `json:"name"`
	// The corresponding value to the product option name.
	Values []*OptionValue `json:"values"`
}

type OptionValue struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
