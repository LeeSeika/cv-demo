package comparingnilinterface

import "testing"

func TestNormalCase(t *testing.T) {
	var productObj = &ProductObject{
		ID:     "prod_123",
		Name:   "Sample Product",
		Handle: "sample-product",
		Price:  19.99,
	}
	model := ObjectToModel(productObj)
	if model == nil {
		t.Errorf("expected non-nil model, got nil")
	}
	switch m := model.(type) {
	case *ProductModel:
		if m.ID != productObj.ID || m.Name != productObj.Name || m.Handle != productObj.Handle || m.Price != productObj.Price {
			t.Errorf("model fields do not match object fields")
		}
	default:
		t.Errorf("expected ProductModel type, got %T", m)
	}
}

func TestNilCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ObjectToModel panicked on nil input: %v", r)
		}
	}()
	var productObj *ProductObject = nil
	model := ObjectToModel(productObj)
	if model != nil {
		t.Errorf("expected nil model, got non-nil")
	}
}
