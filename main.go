package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	http.HandleFunc("/recipe", headers)
	http.ListenAndServe(":9000", nil)
}

func parse_page(url string) string {
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
	str := ""
	var ingredient func(*html.Node)
	ingredient = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "ul" {
			for _, attribute := range n.Attr {
				if attribute.Val == "wprm-recipe-ingredients" {
					str += navigateUlElement(n)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ingredient(c)
		}
	}

	ingredient(doc)
	return str
}

func navigateUlElement(n *html.Node) string {
	str := ""
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == "li" {
			for _, attribute := range n.Attr {
				if attribute.Val == "wprm-recipe-ingredient" {
					str += navigateLiElement(n)
					str += "\n"
				}
			}
		}
	}
	return str
}

func navigateLiElement(n *html.Node) string {
	str := ""
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && (n.Data == "span" || n.Data == "a") {
			str += navigateLiElement(n)
		}
		if n.Type == html.TextNode {
			str += string(n.Data)
		}
	}
	return str
}

func headers(w http.ResponseWriter, req *http.Request) {

	// fmt.Fprintf(w, "%v\n", req.URL.Query().Get("url"))
	param := req.URL.Query().Get("url")
	fmt.Fprintf(w, "%v\n", param)

	if param != "" {
		r := regexp.MustCompile("(\\d+[\\.\\/]*\\d* [a-zA-Z.]*|\\d+)? ([A-Za-z\\d-(),'*. ]+) \\(([\\$0-9.]+)\\)")
		testing := parse_page(param)
		fmt.Print("recipe incoming!\n\n")
		matches := r.FindAllStringSubmatch(testing, -1)

		for _, match := range matches {
			// original match, qty, ingredient, price
			if len(match) != 4 {
				fmt.Println(fmt.Errorf("[!] line not parsed correctly. {%s}", strings.Join(match, "")))
			}
			fmt.Printf("[qty: %s, ingredient: %s, price: %s]\n", match[1], match[2], match[3])
		}
		fmt.Println()
	}
}
