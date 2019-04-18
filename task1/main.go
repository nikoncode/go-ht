package main

func Filter(array []int, predicate func(int, int) bool) (result []int) {
	for index, element := range array {
		if predicate(element, index) {
			result = append(result, element)
		}
	}
	return result
}

func main() {

}
