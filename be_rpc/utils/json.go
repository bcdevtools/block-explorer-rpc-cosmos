package utils

import "encoding/json"

func TryConvertAnyStructToMap(source any) (res map[string]any, err error) {
	res = make(map[string]any)

	var bz []byte
	if source != nil {
		bz, err = json.Marshal(source)
		if err == nil {
			err = json.Unmarshal(bz, &res)
		}
	}

	return
}

func TryGetMapValueAsType[T any](m map[string]any, key string) (value T, ok bool) {
	v, ok := m[key]
	if ok {
		value, ok = v.(T)
	}

	return
}
