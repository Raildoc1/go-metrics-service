package dbrepository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBRepository_formatValues(t *testing.T) {
	tests := []struct {
		name                                string
		firstNumber, valuesCount, rowsCount int
		expectedResult                      string
	}{
		{
			name:           "right format 1, 2*3",
			firstNumber:    1,
			valuesCount:    2,
			rowsCount:      3,
			expectedResult: "($1,$2),($3,$4),($5,$6)",
		},
		{
			name:           "right format 1, 3*2",
			firstNumber:    1,
			valuesCount:    3,
			rowsCount:      2,
			expectedResult: "($1,$2,$3),($4,$5,$6)",
		},
		{
			name:           "right format 15, 2*3",
			firstNumber:    15,
			valuesCount:    2,
			rowsCount:      3,
			expectedResult: "($15,$16),($17,$18),($19,$20)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := formatValuesRows(tt.firstNumber, tt.valuesCount, tt.rowsCount)
			assert.Equal(t, tt.expectedResult, res)
		})
	}
}
