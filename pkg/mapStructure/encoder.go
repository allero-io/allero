package mapStructure

import "github.com/mitchellh/mapstructure"

func Encode(v interface{}) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := mapstructure.Decode(v, &m); err != nil {
		return nil, err
	}
	expandRemainderValues(m)
	return m, nil
}

func expandRemainderValues(m map[string]interface{}) {
	for k, v := range m {
		v, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if k == "" {
			for remainderK, remainderV := range v {
				m[remainderK] = remainderV
			}
			delete(m, "")
		} else {
			expandRemainderValues(v)
		}
	}
}
