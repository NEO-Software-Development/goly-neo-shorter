package utils

import "math/rand"


// runes is a slice of runes used to generate random strings.
var runes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomURL generates a random string of a given size.
// It takes an integer size as a parameter.
// It returns the generated random string.
func RandomURL(size int) string {
	str := make([]rune, size)

	for i := range str {
		str[i] = runes[rand.Intn(len(runes))]
	}

	return string(str)
}