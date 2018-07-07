package challenges

import (
	"database/sql/driver"
	"encoding/json"
)

// we need these to implement scanners and valuers to convert types for db storage

func (c CaseList) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(b), nil
}

func (c *CaseList) Scan(value interface{}) error {
	if value == nil {
		*c = CaseList{}
		return nil
	}

	if bv, err := driver.String.ConvertValue(value); err == nil {
		// if this is a string type
		if v, ok := bv.([]uint8); ok {
			// set the value of the pointer yne to YesNoEnum(v)
			err := json.Unmarshal([]byte(v), c)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func (t tagList) Value() (driver.Value, error) {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(b), nil
}

func (t *tagList) Scan(value interface{}) error {
	if value == nil {
		*t = tagList{}
		return nil
	}

	if bv, err := driver.String.ConvertValue(value); err == nil {
		// if this is a string type
		if v, ok := bv.([]uint8); ok {
			// set the value of the pointer yne to YesNoEnum(v)
			err := json.Unmarshal([]byte(v), t)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}
