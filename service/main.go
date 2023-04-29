package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {

	http.HandleFunc("/recipe", headers)
	http.ListenAndServe(":9000", nil)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		fmt.Println("[!!] Error: Unable to read authorization code", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		fmt.Println("[!!] Error: Unable to retrieve token from web", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("[!!] Error: Unable to cache oauth token", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func parse_page(url string) string {
	fmt.Printf("Getting url {%s} ... \n", url)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("[!!] Error: HTTP Get request failed", err)
	}

	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	if err != nil {
		fmt.Println("[!!] Error: Parsing HTML of website failed", err)
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
	param := req.URL.Query().Get("url")
	u, err := url.Parse(param)
	if err != nil {
		fmt.Println("[!!] Error: Parsing URL failed", err)
	}

	title := strings.Title(strings.ReplaceAll(u.Path[1:len(u.Path)-1], "-", " "))
	if param != "" {
		r := regexp.MustCompile("(\\d+[\\.\\/]*\\d* [a-zA-Z.]*|\\d+)? ([A-Za-z\\d-(),'*. ]+) \\(([\\$0-9.]+)\\)")
		testing := parse_page(param)
		fmt.Print("recipe incoming!\n\n")

		ctx := context.Background()
		b, err := os.ReadFile("credentials.json")
		if err != nil {
			fmt.Println("[!!] Error: Unable to read client secret file", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
		if err != nil {
			fmt.Println("[!!] Error: Unable to parse client secret file to config", err)
		}
		client := getClient(config)

		srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			fmt.Println("[!!] Error: Unable to retrieve Sheets client", err)
		}

		spreadsheetId := "1929iNShf_p-H3QFgUq4xofSnvTLd62qGp8Csi_R9Rbc"
		// readRange := "Sheet1!A2:D"
		// resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
		// if err != nil {
		// 		log.Fatalf("Unable to retrieve data from sheet: %v", err)
		// }

		// if len(resp.Values) == 0 {
		// 		fmt.Println("No data found.")
		// } else {
		// 		fmt.Println("Amount, Ingredient:")
		// 		for _, row := range resp.Values {
		// 			if (len(row) > 1) {
		// 				// Print columns A and E, which correspond to indices 0 and 4.
		// 				fmt.Printf("%s, %s\n", row[0], row[1])
		// 			}
		// 			if (len(row) == 0) {
		// 				fmt.Printf("\n")
		// 			}
		// 		}
		// }
		req := sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: title,
				},
			},
		}

		rbb := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{&req},
		}

		_, err = srv.Spreadsheets.BatchUpdate(spreadsheetId, rbb).Context(ctx).Do()
		if err != nil {
			fmt.Printf("[!] Sheet with title {%s} already exists!\n", title)
		}

		writeRange := title + "!A1"
		var vr sheets.ValueRange
		myval := []interface{}{"Amount", "Ingredient", "Price", "Purchased"}
		vr.Values = append(vr.Values, myval)
		_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").Do()
		if err != nil {
			fmt.Println("[!!] Error: Unable to retrieve data from sheet", err)
		}

		matches := r.FindAllStringSubmatch(testing, -1)

		for idx, match := range matches {
			// original match, qty, ingredient, price
			if len(match) != 4 {
				fmt.Println(fmt.Errorf("[!] line not parsed correctly. {%s}", strings.Join(match, "")))
			}

			writeRange := title + "!A" + strconv.Itoa(idx+2)
			var vr sheets.ValueRange
			myval := []interface{}{match[1], match[2], match[3], ""}
			vr.Values = append(vr.Values, myval)
			_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").Do()
			if err != nil {
				fmt.Println("[!!] Error: Unable to retrieve data from sheet", err)
			}
		}
	}
}
