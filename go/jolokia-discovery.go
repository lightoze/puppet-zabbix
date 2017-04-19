package main

import (
	"os"
	"fmt"
	"strings"
	"net/http"
	"time"
	"encoding/json"
	"reflect"
	"io/ioutil"
	"bytes"
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
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not fully read response body", err)
		return nil
	}

	var results map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	err = decoder.Decode(&results)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse Jolokia response", string(body), err)
		return nil
	}

	values := results["value"]
	escaper := strings.NewReplacer("/", "!/")
	if values != nil && reflect.TypeOf(values).Kind() == reflect.Slice {
		ret := []string{}
		for _, value := range values.([]interface{}) {
			ret = append(ret, escaper.Replace(value.(string)))
		}
		return ret
	} else {
		fmt.Fprintln(os.Stderr, "Incorrect Jolokia search results", results)
		return nil
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Too few arguments")
		os.Exit(1)
	}

	url := os.Args[1]
	query := NormalizePath(ZabbixUnescape(strings.Join(NonEmpty(os.Args[2:]), ",")))

	items := JolokiaSearch(url + "/search/" + query)
	if items == nil {
		os.Exit(1)
	}

	data := []interface{}{}
	for _, item := range items {
		values := make(map[string]string)
		values["{#JMXOBJ}"] = ZabbixEscape(NormalizePath(item))

		domain, path := SplitTwo(item, ":", "")
		path, _ = SplitTwo(path, "/", "!")
		values["{#JMXDOMAIN}"] = domain

		for _, tag := range strings.Split(path, ",") {
			parts := strings.SplitN(tag, "=", 2)
			if len(parts) > 1 {
				values["{#" + strings.ToUpper(parts[0]) + "}"] = parts[1]
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
