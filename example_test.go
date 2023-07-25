package xhttpclient

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ExampleXClient_Do() {
	type Response struct {
		Args struct {
			Hello string `json:"hello"`
		} `json:"args"`
	}

	cli := NewClient()

	var respHello Response
	_, respBody, err := cli.Do(&respHello, nil,
		NewGet().
			Path("https://httpbin.org/get").
			SetQuery("hello", "world"),
	)
	if err != nil {
		log.Fatalf("%s\n%s", err, respBody)
	}

	fmt.Println(respHello.Args.Hello)
	// Output:
	// world
}

func ExampleXClient_Do_unexpected_error() {
	cli := NewClient()

	var respEmpty any
	_, _, err := cli.Do(&respEmpty, nil,
		NewGet().
			Path("https://httpbin.org/status/", "404"),
	)
	fmt.Println(err)
	// Output:
	// unexpected error: URL: https://httpbin.org/status/404 (404 Not Found)
}

func ExampleXClient_Do_wrong_resp() {
	type GitHubError struct {
		Message          string `json:"message"`
		DocumentationUrl string `json:"documentation_url"`
	}

	cli := NewClient().BaseURL("https://api.github.com")

	var (
		respEmpty any
		respWrong GitHubError
	)
	_, respBody, err := cli.Do(&respEmpty, &respWrong,
		NewGet().
			Path("/markdown"),
	)
	if err != nil {
		log.Fatalf("%s\n%s", err, respBody)
	}
	fmt.Println(respWrong.Message)
	fmt.Println(respWrong.DocumentationUrl)
	fmt.Println(string(respBody))
	// Output:
	// Not Found
	// https://docs.github.com/rest
	// {"message":"Not Found","documentation_url":"https://docs.github.com/rest"}
}

func ExampleXClient_Do_upload() {
	tmpTestdata := filepath.Join(os.TempDir(), "testdata.json")
	if err := os.WriteFile(tmpTestdata, []byte(`{"x","go"}`), 0644); err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(tmpTestdata)

	type Response struct {
		Files struct {
			Testfile string `json:"testfile"`
		} `json:"files"`
		Form struct {
			Hello string `json:"hello"`
		} `json:"form"`
	}

	cli := NewClient().BaseURL("https://httpbin.org")

	var resp Response
	_, respBody, err := cli.DoOnceWithBodyCodec(BodyCodecMultipart, &resp, nil,
		NewPost().
			Path("/post").
			Body(
				NewMultipartWriter().
					WriteWithFieldValue("hello", "world").
					WriteWithFile("testfile", tmpTestdata),
			),
	)
	if err != nil {
		log.Fatalf("%s\n%s", err, respBody)
	}

	fmt.Println(resp.Form.Hello)
	fmt.Println(resp.Files.Testfile)
	// Output:
	// world
	// {"x","go"}
}

func ExampleXClient_Do_download() {
	cli := NewClient()

	_, resp, cancel, err := cli.DoWithRaw(
		NewGet().
			Path("https://raw.githubusercontent.com/electricbubble/xhttpclient/main/LICENSE"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cancel()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Body); err != nil {
		log.Fatal(err)
	}

	license, _, _ := strings.Cut(buf.String(), "\n")
	fmt.Println(license)
	// Output:
	// MIT License
}
