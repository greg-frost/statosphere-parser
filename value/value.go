package value

import (
	"fmt"

	"statosphere/parser/check"
	"statosphere/parser/format"
)

// Значение
type Value struct {
	Exact  uint   `json:"exact,omitempty"`
	Approx uint   `json:"approx,omitempty"`
	Short  string `json:"short,omitempty"`
}

// Новое значение
func New(value string, isExact bool) (Value, error) {
	var (
		exact, approx uint
		short         string
	)

	intValue, err := check.PositiveNumber(value)
	if err != nil {
		return Value{}, err
	}
	strValue := format.Trim(value)

	switch isExact || fmt.Sprint(intValue) == strValue {
	case true:
		exact = intValue
	case false:
		approx = intValue
		short = strValue
	}

	return Value{Exact: exact, Approx: approx, Short: short}, nil
}

// Получение значения
func (v Value) Value() uint {
	if v.Exact != 0 {
		return v.Exact
	}

	return v.Approx
}

// Является ли значение точным
func (v Value) IsExact() bool {
	if v.Exact != 0 {
		return true
	}

	return false
}

// Приведение значения к строке
func (v Value) String() string {
	var res string

	if v.Exact != 0 {
		res = fmt.Sprint(v.Exact)
	} else {
		res = fmt.Sprint(v.Approx)
	}

	if v.Short != "" && v.Short != res {
		res += fmt.Sprintf(" (%v)", v.Short)
	}

	return res
}

// Печать значения
func (v Value) Print(title string) {
	fmt.Println(title, v)
}
