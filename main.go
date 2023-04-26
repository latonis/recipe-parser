package main

import (
	"fmt"
	"net/http"
	"strings"
	"regexp"
	"golang.org/x/net/html"
)

func main() {
	r := regexp.MustCompile("(\\d+[\\.\\/]*\\d* [a-zA-Z.]+) ([A-Za-z\\d-()'* ]+) \\(([\\$0-9.]+)\\)")
	testData := `
	12 oz. shrimp (41-60 size) ($4.99)
	1 fresh lemon ($0.60)
	4 cloves garlic ($0.32)
	2 Tbsp butter ($0.24)
	1.5 cups long grain white rice ($0.93)
	2 cups chicken broth ($0.26)
	1/2 cup water ($0.00)
	1 tsp Tony Chachere's seasoning* ($0.10)
	2 Tbsp chopped parsley ($0.09)
	`
	fmt.Print("recipe incoming!\n\n")
	matches := r.FindAllStringSubmatch(testData, -1)
	
	for _, match := range matches {
		// original match, qty, ingredient, price
		if len(match) != 4 {
			fmt.Println(fmt.Errorf("[!] line not parsed correctly. {%s}", strings.Join(match, "")))
		}
		fmt.Printf("[qty: %s, ingredient: %s, price: %s]\n", match[1], match[2], match[3])
	}
	fmt.Println()
	parse_page("https://www.budgetbytes.com/sriracha-egg-salad/")
}

func parse_page(url string) {
	fmt.Printf("Getting url {%s} ... \n", url)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("[!!] Error: ", err)
	}
	
	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	if err != nil {
		fmt.Println("[!!] Error: ", err)
	}

	// var ingredients []string
	var ingredient func (*html.Node)
	ingredient = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			for _, attribute := range n.Attr {
				if (attribute.Val == "wprm-recipe-ingredient") {
					fmt.Println(attribute)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ingredient(c)
		}
	}
	ingredient(doc)
}