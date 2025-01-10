package main

import "fmt"

func main() {
	for i := range 8 {
		fmt.Printf("\x1b[%dm███\x1b[0m", 30+i)
	}
	fmt.Println()
	for i := range 8 {
		fmt.Printf("\x1b[%dmABC\x1b[0m", 30+i)
	}
	fmt.Println()
	for i := range 8 {
		fmt.Printf("\x1b[1;%dm███\x1b[0m", 30+i)
	}
	fmt.Println()
	for i := range 8 {
		fmt.Printf("\x1b[1;%dmABC\x1b[0m", 30+i)
	}
	fmt.Println()
	for i := range 8 {
		fmt.Printf("\x1b[%dm   \x1b[0m", 40+i)
	}
	fmt.Println()
	for i := range 8 {
		fmt.Printf("\x1b[1;%dm   \x1b[0m", 40+i)
	}
	fmt.Println()
}
