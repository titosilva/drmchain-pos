package clru

type clruNode[K comparable, V any] struct {
	key   K
	value V
	next  *clruNode[K, V]
	prev  *clruNode[K, V]
}
