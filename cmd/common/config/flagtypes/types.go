package flagtypes

import (
	"flag"
	"strconv"
)

const nilString = "<nil>"

// String.

var _ flag.Value = (*String)(nil)

type String struct {
	Argument[string]
}

func NewString() *String {
	return &String{
		*newArgument[string](),
	}
}

func (s *String) String() string {
	if s.val == nil {
		return nilString
	}
	return *s.val
}

func (s *String) Set(input string) error {
	s.val = &input
	return nil
}

// Int.

var _ flag.Value = (*Int)(nil)

type Int struct {
	Argument[int]
}

func NewInt() *Int {
	return &Int{
		*newArgument[int](),
	}
}

func (s *Int) String() string {
	if s.val == nil {
		return nilString
	}
	return strconv.Itoa(*s.val)
}

func (s *Int) Set(input string) error {
	val, err := strconv.Atoi(input)
	if err != nil {
		return err
	}
	s.val = &val
	return nil
}
