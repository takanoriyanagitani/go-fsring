package fsring

func ManagerUintMemNew[T uint8 | uint16](init T) ManagerUint[T] {
	var t T = init
	var get GetUint[T] = func() (T, error) { return t, nil }
	var set SetUint[T] = func(neo T) error {
		t = neo
		return nil
	}
	return ManagerUint[T]{
		get,
		set,
	}
}
