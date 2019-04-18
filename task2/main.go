package main

func MapTo(array []int, mapper func(int, int) string) (result []string) {
	for index, element := range array {
		result = append(result, mapper(element, index))
	}
	return result
}

//Kind of class level constant to avoid wasting memory
var numToText = []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

func Convert(array []int) []string {
	/*
		Can be also here, doesn't matter. What do you prefer?
		var numToText = []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	*/
	return MapTo(array, func(element, index int) string {
		if element < 1 || element > 9 {
			return "unknown"
		}
		return numToText[element-1]
	})
}

func main() {
}
