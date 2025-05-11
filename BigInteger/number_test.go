package uint239

import (
	"testing"
)

func TestFromUint32(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint32
		shift    uint32
		expected Uint239
	}{
		{
			name:     "zero_no_shift",
			value:    0,
			shift:    0,
			expected: Uint239{},
		},
		{
			name:  "one_no_shift",
			value: 1,
			shift: 0,
			expected: func() Uint239 {
				result := Uint239{}
				result.Data[34] = 1
				return result
			}(),
		},
		{
			name:  "small_with_shift",
			value: 42,
			shift: 7,
			expected: func() Uint239 {

				result := Uint239{}

				result.Data[0] = 0x80
				result.Data[1] = 0x80
				result.Data[2] = 0x80

				return result
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FromUint32(tc.value, tc.shift)

			if tc.value <= 100 && tc.shift == 0 {
				rawValue := toUint32(result)
				if rawValue != tc.value {
					t.Errorf("Expected value %d, got %d", tc.value, rawValue)
				}
			}

			shift := GetShift(result)
			if shift != tc.shift {
				t.Errorf("Expected shift %d, got %d", tc.shift, shift)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		shift    uint32
		expected uint32
	}{
		{
			name:     "zero",
			value:    "0",
			shift:    0,
			expected: 0,
		},
		{
			name:     "small_number",
			value:    "42",
			shift:    0,
			expected: 42,
		},
		{
			name:     "small_with_shift",
			value:    "123",
			shift:    5,
			expected: 123,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FromString(tc.value, tc.shift)

			rawValue := toUint32(removeShift(result))
			if rawValue != tc.expected {
				t.Errorf("Expected value %d, got %d", tc.expected, rawValue)
			}

			shift := GetShift(result)
			if shift != tc.shift {
				t.Errorf("Expected shift %d, got %d", tc.shift, shift)
			}
		})
	}
}

func TestArithmeticOperations(t *testing.T) {

	t.Run("Addition", func(t *testing.T) {

		a := FromUint32(10, 0)
		b := FromUint32(20, 0)
		sum := Add(a, b)

		if toUint32(removeShift(sum)) != 30 {
			t.Errorf("Expected 10 + 20 = 30, got %d", toUint32(removeShift(sum)))
		}

		if GetShift(sum) != 0 {
			t.Errorf("Expected shift 0, got %d", GetShift(sum))
		}

		a = FromUint32(10, 3)
		b = FromUint32(20, 5)
		sum = Add(a, b)

		if toUint32(removeShift(sum)) != 30 {
			t.Errorf("Expected 10 + 20 = 30, got %d", toUint32(removeShift(sum)))
		}

		if GetShift(sum) != 8 {
			t.Errorf("Expected shift 8, got %d", GetShift(sum))
		}
	})

	t.Run("Subtraction", func(t *testing.T) {

		a := FromUint32(30, 0)
		b := FromUint32(10, 0)
		diff := Subtract(a, b)

		if toUint32(removeShift(diff)) != 20 {
			t.Errorf("Expected 30 - 10 = 20, got %d", toUint32(removeShift(diff)))
		}

		if GetShift(diff) != 0 {
			t.Errorf("Expected shift 0, got %d", GetShift(diff))
		}

		a = FromUint32(30, 7)
		b = FromUint32(10, 2)
		diff = Subtract(a, b)

		if toUint32(removeShift(diff)) != 20 {
			t.Errorf("Expected 30 - 10 = 20, got %d", toUint32(removeShift(diff)))
		}

		if GetShift(diff) != 5 {
			t.Errorf("Expected shift 5, got %d", GetShift(diff))
		}
	})

	t.Run("Multiplication", func(t *testing.T) {

		a := FromUint32(6, 0)
		b := FromUint32(7, 0)
		prod := Multiply(a, b)

		if toUint32(removeShift(prod)) != 42 {
			t.Errorf("Expected 6 * 7 = 42, got %d", toUint32(removeShift(prod)))
		}

		if GetShift(prod) != 0 {
			t.Errorf("Expected shift 0, got %d", GetShift(prod))
		}

		a = FromUint32(6, 3)
		b = FromUint32(7, 4)
		prod = Multiply(a, b)

		if toUint32(removeShift(prod)) != 42 {
			t.Errorf("Expected 6 * 7 = 42, got %d", toUint32(removeShift(prod)))
		}

		if GetShift(prod) != 7 {
			t.Errorf("Expected shift 7, got %d", GetShift(prod))
		}
	})

	t.Run("Division", func(t *testing.T) {

		a := FromUint32(42, 0)
		b := FromUint32(6, 0)
		quot := Divide(a, b)

		if toUint32(removeShift(quot)) != 7 {
			t.Errorf("Expected 42 / 6 = 7, got %d", toUint32(removeShift(quot)))
		}

		if GetShift(quot) != 0 {
			t.Errorf("Expected shift 0, got %d", GetShift(quot))
		}

		a = FromUint32(42, 10)
		b = FromUint32(6, 3)
		quot = Divide(a, b)

		if toUint32(removeShift(quot)) != 7 {
			t.Errorf("Expected 42 / 6 = 7, got %d", toUint32(removeShift(quot)))
		}

		if GetShift(quot) != 7 {
			t.Errorf("Expected shift 7, got %d", GetShift(quot))
		}
	})
}

func TestEquality(t *testing.T) {

	t.Run("Equal_no_shift", func(t *testing.T) {
		a := FromUint32(123, 0)
		b := FromUint32(123, 0)

		if !Equal(a, b) {
			t.Errorf("Expected 123 == 123 to be true")
		}

		if NotEqual(a, b) {
			t.Errorf("Expected 123 != 123 to be false")
		}
	})

	t.Run("Equal_with_shift", func(t *testing.T) {
		a := FromUint32(123, 5)
		b := FromUint32(123, 10)

		if !Equal(a, b) {
			t.Errorf("Expected 123 (shift 5) == 123 (shift 10) to be true")
		}

		if NotEqual(a, b) {
			t.Errorf("Expected 123 (shift 5) != 123 (shift 10) to be false")
		}
	})

	t.Run("Not_equal", func(t *testing.T) {
		a := FromUint32(123, 0)
		b := FromUint32(456, 0)

		if Equal(a, b) {
			t.Errorf("Expected 123 == 456 to be false")
		}

		if !NotEqual(a, b) {
			t.Errorf("Expected 123 != 456 to be true")
		}
	})
}
