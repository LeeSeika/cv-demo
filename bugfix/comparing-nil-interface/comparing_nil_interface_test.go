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
}

func TestNilCase_WithBug(t *testing.T) {
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

func TestNilCase_BugFixed(t *testing.T) {
	var productObj *ProductObject = nil
	model := ObjectToModel_Fixed(productObj)
	if model != nil {
		t.Errorf("expected nil model, got non-nil")
	}
}
