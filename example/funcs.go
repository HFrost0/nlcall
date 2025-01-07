package main

import (
	"fmt"
)

func greet(name string, age int) string {
	return fmt.Sprintf("Hello, %s! You are %d years old.", name, age)
}

func add(nums []int) int {
	sum := 0
	for _, num := range nums {
		sum += num
	}
	return sum
}

func no() string {
	return "Sorry, I can't help you with that."
}

// weather is a function that returns the weather in a city
func weather(city string) string {
	return fmt.Sprintf("The weather in %s is sunny.", city)
}

func lengthOfLongestSubstring(s string) int {
	if len(s) <= 1 {
		return len(s)
	}
	m := make(map[byte]struct{})
	i, j := 0, 0
	res := 0
	m[s[i]] = struct{}{}
	m[s[j]] = struct{}{}
	for j < len(s)-1 {
		if _, ok := m[s[j+1]]; !ok {
			j += 1
			res = max(res, j-i+1)
			m[s[j]] = struct{}{}
		} else {
			delete(m, s[i])
			i += 1
		}
	}
	return res
}

func mul(nums []int) int {
	res := 1
	for _, num := range nums {
		res *= num
	}
	return res
}
