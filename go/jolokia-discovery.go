package main

import (
	"os"
	"fmt"
	"strings"
	"net/http"
	"time"
	"encoding/json"
	"reflect"
)

func JolokiaSearch(url string) []string {
	client := &http.Client{Timeout:3 * time.Second}
	response, err := client.Get(url)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	if response.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "Unexpected HTTP response status", response.StatusCode)
		return nil
	}

	var results map[string]interface{}
	decoder := json.NewDecoder(response.Body)
	decoder.UseNumber()
	err = decoder.Decode(&results)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse Jolokia response", err)
		return nil
	}

	values := results["value"]
	if values != nil && reflect.TypeOf(values).Kind() == reflect.Slice {
		ret := []string{}
		for _, value := range values.([]interface{}) {
			ret = append(ret, value.(string))
		}
		return ret
	} else {
		fmt.Fprintln(os.Stderr, "Incorrect Jolokia search results", results)
		return nil
	}
}

func NonEmpty(args []string) (ret []string) {
	for _, arg := range args {
		if len(arg) > 0 {
			ret = append(ret, arg)
		}
	}
	return
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Too few arguments")
		os.Exit(1)
	}

	url := os.Args[1]
	query := strings.Join(NonEmpty(os.Args[2:]), ",")

	items := JolokiaSearch(url + "/search/" + query)
	if items == nil {
		os.Exit(1)
	}

	data := []interface{}{}
	for _, item := range items {
		values := make(map[string]string)
		values["{#JMXOBJ}"] = item

		base := strings.SplitN(item, ":", 2)
		values["{#JMXDOMAIN}"] = base[0]

		if len(base) > 1 {
			for _, tag := range strings.Split(base[1], ",") {
				parts := strings.SplitN(tag, "=", 2)
				if len(parts) > 1 {
					values["{#" + strings.ToUpper(parts[0]) + "}"] = parts[1]
				}
			}
		}

		data = append(data, values)
	}

	response, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(string(response))
}
