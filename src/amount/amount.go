package amount

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Amount int64

var (
	ErrNegative = errors.New("amount must not be negative")
	ErrOverflow = errors.New("amount overflow")
)

func New(value int64) (Amount, error) {
	if value < 0 {
		return 0, ErrNegative
	}
	return Amount(value), nil
}

func Must(value int64) Amount {
	amount, err := New(value)
	if err != nil {
		panic(err)
	}
	return amount
}

func Zero() Amount {
	return 0
}

func (a Amount) Int64() int64 {
	return int64(a)
}

func (a Amount) String() string {
	return strconv.FormatInt(int64(a), 10)
}

func (a Amount) IsZero() bool {
	return a == 0
}

func (a Amount) IsPositive() bool {
	return a > 0
}

func (a Amount) Validate() error {
	if a < 0 {
		return ErrNegative
	}
	return nil
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*a = 0
		return nil
	}
	var text string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	} else {
		text = string(data)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		*a = 0
		return nil
	}
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", text, err)
	}
	if value < 0 {
		return ErrNegative
	}
	*a = Amount(value)
	return nil
}

func (a Amount) CheckedAdd(other Amount) (Amount, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}
	if err := other.Validate(); err != nil {
		return 0, err
	}
	if int64(a) > math.MaxInt64-int64(other) {
		return 0, ErrOverflow
	}
	return a + other, nil
}

func (a Amount) CheckedSub(other Amount) (Amount, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}
	if err := other.Validate(); err != nil {
		return 0, err
	}
	if other > a {
		return 0, fmt.Errorf("amount underflow: %s < %s", a, other)
	}
	return a - other, nil
}

func (a Amount) SubFloor(other Amount) Amount {
	if other >= a {
		return 0
	}
	return a - other
}

func (a Amount) CheckedMul(value int64) (Amount, error) {
	if err := a.Validate(); err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, ErrNegative
	}
	if value == 0 || a == 0 {
		return 0, nil
	}
	if a != 0 && int64(a) > math.MaxInt64/value {
		return 0, ErrOverflow
	}
	return Amount(int64(a) * value), nil
}

func (a Amount) MulBps(bps int64) (Amount, error) {
	if bps < 0 {
		return 0, ErrNegative
	}
	product, err := a.CheckedMul(bps)
	if err != nil {
		return 0, err
	}
	return Amount(int64(product) / 10_000), nil
}

func (a Amount) MulBpsCeil(bps int64) (Amount, error) {
	if bps < 0 {
		return 0, ErrNegative
	}
	product, err := a.CheckedMul(bps)
	if err != nil {
		return 0, err
	}
	if product == 0 {
		return 0, nil
	}
	return Amount((int64(product) + 9_999) / 10_000), nil
}

func (a Amount) Ratio(numerator, denominator int64) (Amount, error) {
	if numerator < 0 || denominator <= 0 {
		return 0, ErrNegative
	}
	product, err := a.CheckedMul(numerator)
	if err != nil {
		return 0, err
	}
	return Amount(int64(product) / denominator), nil
}

func (a Amount) MustAdd(other Amount) Amount {
	value, err := a.CheckedAdd(other)
	if err != nil {
		panic(err)
	}
	return value
}

func (a Amount) MustSub(other Amount) Amount {
	value, err := a.CheckedSub(other)
	if err != nil {
		panic(err)
	}
	return value
}

func (a Amount) MustMulBps(bps int64) Amount {
	value, err := a.MulBps(bps)
	if err != nil {
		panic(err)
	}
	return value
}

func Sum(values ...Amount) (Amount, error) {
	var total Amount
	for _, value := range values {
		next, err := total.CheckedAdd(value)
		if err != nil {
			return 0, err
		}
		total = next
	}
	return total, nil
}

func Min(a, b Amount) Amount {
	if a < b {
		return a
	}
	return b
}

func Max(a, b Amount) Amount {
	if a > b {
		return a
	}
	return b
}

func Clamp(value, minimum, maximum Amount) Amount {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
