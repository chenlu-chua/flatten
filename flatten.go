// Flatten nested maps or JSON structures to one-dimensional scalar sets.
//
// Intended for JSON APIs, flatten operates on interior maps, slices and scalars.  Flat, compound
// keys are generated in either dotted style (e.g. a.b.1.c) or Rails-like (e.g. a[b][1][c]).
//
// You can flatten JSON strings.
//
//	nested := `{
//	  "one": {
//	    "two": [
//	      "2a",
//	      "2b"
//	    ]
//	  },
//	  "side": "value"
//	}`
//
//	flat, err := FlattenString(nested, "", DOT_STYLE)
//
//	// output: `{ "one.two.0": "2a", "one.two.1": "2b", "side": "value" }`
//
// Or Go maps directly.
//
//	t := map[string]interface{}{
//		"a": "b",
//		"c": map[string]interface{}{
//			"d": "e",
//			"f": "g",
//		},
//		"z": 1.4567
//	}
//
//	flat, err := Flatten(nested, "", RAILS_STYLE)
//
//	// output:
//	// map[string]interface{}{
//	//	"a":    "b",
//	//	"c[d]": "e",
//	//	"c[f]": "g",
//	//	"z":    1.4567,
//	// }
//
package flatten

import (
	"encoding/json"
	"errors"
	"strconv"
)

// The presentation style of keys.
type SeparatorStyle int

const (
	// Separate nested key components with dots, e.g. "a.b.1.c.d"
	DOT_STYLE SeparatorStyle = iota

	// Separate ala Rails, e.g. "a[b][c][1][d]"
	RAILS_STYLE
)

// Nested input must be a map or slice
var NotValidInputError = errors.New("Not a valid input: map or slice")

// Flatten generates a flat map from a nested one.  The original may include values of type map, slice and scalar,
// but not struct.  Keys in the flat map will be a compound of descending map keys and slice iterations.
// The presentation of keys is set by style.  A prefix is joined to each key.
func Flatten(nested map[string]interface{}, prefix string, style SeparatorStyle) (map[string]interface{}, error) {
	flatmap := make(map[string]interface{})

	err := flatten(true, flatmap, nested, prefix, style)
	if err != nil {
		return nil, err
	}

	return flatmap, nil
}

// FlattenString generates a flat JSON map from a nested one.  Keys in the flat map will be a compound of
// descending map keys and slice iterations.  The presentation of keys is set by style.  A prefix is joined
// to each key.
func FlattenString(nestedstr, prefix string, style SeparatorStyle) (string, error) {
	var nested map[string]interface{}
	err := json.Unmarshal([]byte(nestedstr), &nested)
	if err != nil {
		return "", err
	}

	flatmap, err := Flatten(nested, prefix, style)
	if err != nil {
		return "", err
	}

	flatb, err := json.Marshal(&flatmap)
	if err != nil {
		return "", err
	}

	return string(flatb), nil
}

func flatten(top bool, flatMap map[string]interface{}, nested interface{}, prefix string, style SeparatorStyle) error {
	assign := func(newKey string, v interface{}) error {
		switch v.(type) {
		case map[string]interface{}:
			if err := flatten(false, flatMap, v, newKey, style); err != nil {
				return err
			}
		case []interface{}:
			if err := flatten(false, flatMap, v, newKey, style); err != nil {
				return err
			}
		default:
			flatMap[newKey] = v
		}

		return nil
	}

	switch nested.(type) {
	case map[string]interface{}:
		for k, v := range nested.(map[string]interface{}) {
			newKey := enkey(top, prefix, k, style)
			assign(newKey, v)
		}
	case []interface{}:
		for i, v := range nested.([]interface{}) {
			newKey := enkey(top, prefix, strconv.Itoa(i), style)
			assign(newKey, v)
		}
	default:
		return NotValidInputError
	}

	return nil
}

func enkey(top bool, prefix, subkey string, style SeparatorStyle) string {
	key := prefix

	if top {
		key += subkey
	} else {
		switch style {
		case DOT_STYLE:
			key += "." + subkey
		case RAILS_STYLE:
			key += "[" + subkey + "]"
		}
	}

	return key
}