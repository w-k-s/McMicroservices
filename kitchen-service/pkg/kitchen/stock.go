package kitchen

import (
	"fmt"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type StockItem struct {
	name  string
	units uint
}

type StockItemRecord interface {
	Name() string
	Units() uint
}

func NewStockItem(name string, units uint) (StockItem, error) {

	errors := validate.Validate(
		&validators.StringLengthInRange{Name: "Name", Field: name, Min: 1, Max: 25, Message: "Name must be 1 and 25 characters long"},
		&validators.IntIsGreaterThan{Name: "Units", Field: int(units), Compared: 0, Message: "Units must be greater than 0"},
	)

	if err := makeCoreValidationError(ErrInvalidStockItem, errors); err != nil {
		return StockItem{}, err
	}

	return StockItem{
		name,
		units,
	}, nil
}

func NewAccountFromRecord(record StockItemRecord) (StockItem, error) {
	return NewStockItem(record.Name(), record.Units())
}

func (s StockItem) Name() string {
	return s.name
}

func (s StockItem) Units() uint {
	return s.units
}

func (s StockItem) String() string {
	return fmt.Sprintf("StockItem{name: %q, units: %d}", s.name, s.units)
}

type Stock []StockItem
