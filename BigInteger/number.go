package uint239

type Uint239 struct {
	Data [35]byte
}

func FromUint32(value uint32, shift uint32) Uint239 {
	result := Uint239{}

	shiftedValue := make([]byte, 35)

	for i := 0; value > 0 && i < len(shiftedValue); i++ {
		shiftedValue[len(shiftedValue)-1-i] = byte(value & 0x7F)
		value >>= 7
	}

	if shift > 0 {
		shift = shift % 239

		serviceBytes := createServiceBits(shift)

		shiftedValue = circularLeftShift(shiftedValue, shift)

		for i := 0; i < len(shiftedValue); i++ {
			result.Data[i] = shiftedValue[i] | serviceBytes[i]
		}
	} else {

		for i := 0; i < len(shiftedValue); i++ {
			result.Data[i] = shiftedValue[i]
		}
	}

	return result
}

func FromString(str string, shift uint32) Uint239 {
	var value uint64 = 0

	for _, ch := range str {
		if ch >= '0' && ch <= '9' {
			value = value*10 + uint64(ch-'0')
		}
	}

	return FromUint32(uint32(value), shift)
}

func Add(lhs, rhs Uint239) Uint239 {

	lhsShift := GetShift(lhs)
	rhsShift := GetShift(rhs)

	resultShift := lhsShift + rhsShift

	lhsValue := removeShift(lhs)
	rhsValue := removeShift(rhs)

	resultValue := addRaw(lhsValue, rhsValue)

	return applyShift(resultValue, resultShift)
}

func Subtract(lhs, rhs Uint239) Uint239 {

	lhsShift := GetShift(lhs)
	rhsShift := GetShift(rhs)

	resultShift := lhsShift - rhsShift

	lhsValue := removeShift(lhs)
	rhsValue := removeShift(rhs)

	resultValue := subtractRaw(lhsValue, rhsValue)

	return applyShift(resultValue, resultShift)
}

func Multiply(lhs, rhs Uint239) Uint239 {

	lhsShift := GetShift(lhs)
	rhsShift := GetShift(rhs)

	resultShift := lhsShift + rhsShift

	lhsValue := removeShift(lhs)
	rhsValue := removeShift(rhs)

	resultValue := multiplyRaw(lhsValue, rhsValue)

	return applyShift(resultValue, resultShift)
}

func Divide(lhs, rhs Uint239) Uint239 {

	lhsShift := GetShift(lhs)
	rhsShift := GetShift(rhs)

	resultShift := lhsShift - rhsShift

	lhsValue := removeShift(lhs)
	rhsValue := removeShift(rhs)

	resultValue := divideRaw(lhsValue, rhsValue)

	return applyShift(resultValue, resultShift)
}

func Equal(lhs, rhs Uint239) bool {

	lhsValue := removeShift(lhs)
	rhsValue := removeShift(rhs)

	for i := 0; i < len(lhsValue.Data); i++ {

		lhsBits := lhsValue.Data[i] & 0x7F
		rhsBits := rhsValue.Data[i] & 0x7F

		if lhsBits != rhsBits {
			return false
		}
	}

	return true
}

func NotEqual(lhs, rhs Uint239) bool {
	return !Equal(lhs, rhs)
}

func (u Uint239) String() string {

	return "Implement me"
}

func GetShift(value Uint239) uint32 {
	var shift uint32

	for i := 0; i < len(value.Data); i++ {

		if value.Data[i]&0x80 != 0 {

			shift |= 1 << i
		}
	}

	return shift
}

func createServiceBits(shift uint32) [35]byte {
	var result [35]byte

	for i := 0; i < 35 && shift > 0; i++ {
		if shift&1 != 0 {
			result[i] = 0x80
		}
		shift >>= 1
	}

	return result
}

func circularLeftShift(value []byte, shift uint32) []byte {
	result := make([]byte, len(value))

	byteShift := shift / 7
	bitShift := shift % 7

	for i := 0; i < len(value); i++ {

		srcIdx := (i - int(byteShift) + len(value)) % len(value)

		srcBits := value[srcIdx] & 0x7F

		if bitShift > 0 {

			prevIdx := (srcIdx - 1 + len(value)) % len(value)
			prevBits := value[prevIdx] & 0x7F

			result[i] = (srcBits << bitShift) | (prevBits >> (7 - bitShift))
			result[i] &= 0x7F
		} else {

			result[i] = srcBits
		}
	}

	return result
}

func removeShift(value Uint239) Uint239 {
	shift := GetShift(value)
	result := Uint239{}

	for i := 0; i < len(value.Data); i++ {

		result.Data[i] = value.Data[i] & 0x7F
	}

	if shift > 0 {

		valueBytes := make([]byte, 35)
		for i := 0; i < 35; i++ {
			valueBytes[i] = result.Data[i]
		}

		valueBytes = circularRightShift(valueBytes, shift)

		for i := 0; i < 35; i++ {
			result.Data[i] = valueBytes[i]
		}
	}

	return result
}

func circularRightShift(value []byte, shift uint32) []byte {

	totalBits := uint32(239)
	return circularLeftShift(value, totalBits-shift%totalBits)
}

func applyShift(value Uint239, shift uint32) Uint239 {

	valueBytes := make([]byte, 35)
	for i := 0; i < 35; i++ {
		valueBytes[i] = value.Data[i]
	}

	shiftedBytes := circularLeftShift(valueBytes, shift)

	serviceBits := createServiceBits(shift)

	result := Uint239{}
	for i := 0; i < 35; i++ {
		result.Data[i] = shiftedBytes[i] | serviceBits[i]
	}

	return result
}

func addRaw(lhs, rhs Uint239) Uint239 {
	result := Uint239{}
	var carry byte

	for i := len(lhs.Data) - 1; i >= 0; i-- {

		sum := (lhs.Data[i] & 0x7F) + (rhs.Data[i] & 0x7F) + carry

		result.Data[i] = sum & 0x7F

		carry = sum >> 7
	}

	return result
}

func subtractRaw(lhs, rhs Uint239) Uint239 {
	result := Uint239{}
	var borrow byte

	for i := len(lhs.Data) - 1; i >= 0; i-- {
		lhsByte := lhs.Data[i] & 0x7F
		rhsByte := rhs.Data[i] & 0x7F

		if lhsByte >= rhsByte+borrow {
			result.Data[i] = lhsByte - (rhsByte + borrow)
			borrow = 0
		} else {

			result.Data[i] = 0x80 + lhsByte - (rhsByte + borrow)
			borrow = 1
		}

		result.Data[i] &= 0x7F
	}

	return result
}

func multiplyRaw(lhs, rhs Uint239) Uint239 {

	lhsVal := toUint32(lhs)
	rhsVal := toUint32(rhs)

	product := lhsVal * rhsVal

	return rawFromUint32(product)
}

func divideRaw(lhs, rhs Uint239) Uint239 {

	lhsVal := toUint32(lhs)
	rhsVal := toUint32(rhs)

	var quotient uint32
	if rhsVal != 0 {
		quotient = lhsVal / rhsVal
	}

	return rawFromUint32(quotient)
}

func toUint32(value Uint239) uint32 {
	var result uint32

	for i := 0; i < len(value.Data); i++ {

		bits := value.Data[i] & 0x7F
		result = (result << 7) | uint32(bits)
	}

	return result
}

func rawFromUint32(value uint32) Uint239 {
	result := Uint239{}

	for i := len(result.Data) - 1; i >= 0 && value > 0; i-- {
		result.Data[i] = byte(value & 0x7F)
		value >>= 7
	}

	return result
}
