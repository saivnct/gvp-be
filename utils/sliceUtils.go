package utils

func RemoveSliceDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func IsSlicesEqual[T string | int](sliceList1 []T, sliceList2 []T) bool {
	if len(sliceList1) != len(sliceList2) {
		return false
	}

	for index, item := range sliceList1 {
		if item != sliceList2[index] {
			return false
		}
	}

	return true
}
