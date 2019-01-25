package colorful

import (
	"database/sql/driver"
	"fmt"
	"reflect"
)

// A HexColor is a Color stored as a hex string "#rrggbb". It implements the
// database/sql.Scanner and database/sql/driver.Value interfaces.
type HexColor Color

type errUnsupportedType struct {
	got  interface{}
	want reflect.Type
}

func (hc *HexColor) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errUnsupportedType{got: reflect.TypeOf(value), want: reflect.TypeOf("")}
	}
	c, err := Hex(s)
	if err != nil {
		return err
	}
	*hc = HexColor(c)
	return nil
}

func (hc *HexColor) Value() (driver.Value, error) {
	return Color(*hc).Hex(), nil
}

func (e errUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: got %v, want a %s", e.got, e.want)
}
