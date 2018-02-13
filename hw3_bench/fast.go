package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	if err != nil {
		panic(err)
	}

	seenBrowsers := make(map[string]bool)

	index := -1
	user := &User{}
	var isAndroid, isMSIE bool
	var email string

	fmt.Fprintln(out, "found users:")
	for scanner.Scan() {
		index++
		err := json.Unmarshal(scanner.Bytes(), &user)
		if err != nil {
			panic(err)
		}

		isAndroid = false
		isMSIE = false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				seenBrowsers[browser] = true
			} else if strings.Contains(browser, "MSIE") {
				isMSIE = true
				seenBrowsers[browser] = true
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email = strings.Replace(user.Email, "@", " [at] ", -1)
		fmt.Fprintf(out, "[%d] %s <%s>\n", index, user.Name, email)
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
