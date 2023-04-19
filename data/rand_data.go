package data

import "math/rand"

// Package for generating random data

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Each goroutine client should pass in their own RNG
func RandInt(rng *rand.Rand, start, stop int) int {
	return rng.Intn(stop) + start
}

// Each goroutine client should pass in their own RNG
func RandComment(rng *rand.Rand, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}

// Each goroutine client should pass in their own RNG
func RandDirection(rng *rand.Rand) string {
	if rng.Intn(2) == 1 { // flip a coin
		return "right"
	}
	return "left"
}
