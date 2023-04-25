package main

import (
	"fmt"
	// "bytes"
	"regexp"
)

func main() {
	r, _ := regexp.Compile("(\\d+[\\.\\/]*\\d* [a-zA-Z.]+) ([A-Za-z\\d-()'* ]+) (\\([\\$0-9.]+\\))")
	fmt.Println("recipe incoming!")

	fmt.Println(r.MatchString("12 oz. shrimp (41-60 size) ($4.99)"))
	fmt.Print(r.MatchString("1 tsp Tony Chachere's seasoning* ($0.10)"))
}
