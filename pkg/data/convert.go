package data

func Slice(s Iterable) []interface{} {
	a := []interface{}{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next())
	}
	return a
}

func StringSlice(s Iterable) []string {
	a := []string{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next().(string))
	}
	return a
}

func StringArraySlice(s Iterable) [][]string {
	a := [][]string{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next().([]string))
	}
	return a
}