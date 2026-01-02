package template

import (
	"fmt"

	"github.com/leeseika/cv-demo/pkg/jsonx"
	"github.com/leeseika/cv-demo/pkg/page/material/component/element"
)

type (
	ElementValueHandler interface {
		Handle(element element.Element, val jsonx.JSONValue, prevErr error) (jsonx.JSONValue, error)
	}

	elementValueChecker       struct{}
	elementValueDefaultSetter struct{}
)

func NewElementValueChecker() ElementValueHandler {
	return &elementValueChecker{}
}

func NewElementValueDefaultSetter() ElementValueHandler {
	return &elementValueDefaultSetter{}
}

func (evc *elementValueChecker) Handle(ele element.Element, val jsonx.JSONValue, prevErr error) (jsonx.JSONValue, error) {
	if prevErr != nil {
		return val, prevErr
	}

	checkedEleVal, err := ele.CheckValue(val)
	if err != nil {
		return val, fmt.Errorf("component %s element %s value check failed: %w", "", ele.GetID(), err)
	}

	return checkedEleVal, nil
}

func (evds *elementValueDefaultSetter) Handle(ele element.Element, val jsonx.JSONValue, prevErr error) (jsonx.JSONValue, error) {
	if prevErr == nil {
		return val, nil
	}

	defaultVal := ele.GetDefault()
	return defaultVal, nil
}
