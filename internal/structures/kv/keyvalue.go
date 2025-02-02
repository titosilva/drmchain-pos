package kv

type KeyValue[K, V any] struct {
	Key   K
	Value V
}

func New[K, V any](key K, value V) KeyValue[K, V] {
	return KeyValue[K, V]{Key: key, Value: value}
}

func GetKey[K, V any](keyValue KeyValue[K, V]) K {
	return keyValue.Key
}

func GetValue[K, V any](keyValue KeyValue[K, V]) V {
	return keyValue.Value
}
