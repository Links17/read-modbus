package main

import (
	"fmt"
)

func main() {
	xyCoord, err := parseXY("12034")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("x: %d, y: %d\n", xyCoord.X, xyCoord.Y)
	}

	whDim, err := parseWH("34156")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("w: %d, h: %d\n", whDim.W, whDim.H)
	}
}
