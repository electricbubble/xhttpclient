package xhttpclient

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

func ExampleNewClient_query() {
	type Response struct {
		Args struct {
			Hello string `json:"hello"`
		} `json:"args"`
	}

	cli := NewClient()

	r := NewGet().
		Path("https://httpbin.org/get").
		QuerySet("hello", "world")
	var success Response
	resp, err := cli.Do(r, CodecJSON, &success, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()

	fmt.Println(success.Args.Hello)
	// Output:
	// world
}

func ExampleNewClient_notFound() {
	cli := NewClient()

	r := NewDelete().
		Path("https://httpbin.org/status/", "404")
	_, err := cli.Do(r, CodecJSON, nil, nil)
	fmt.Println(err)
	// Output:
	// decode: unhandled
	// Method: DELETE
	// URL: https://httpbin.org/status/404
	// StatusCode: 404
	// Body(Len: 0):
}

func ExampleNewClient_failure() {
	type GitHubError struct {
		Message          string `json:"message"`
		DocumentationUrl string `json:"documentation_url"`
		Status           string `json:"status"`
	}

	cli := NewClient().
		BaseURL("https://api.github.com")

	r := NewGet().
		Path("/markdown")
	var failure GitHubError
	resp, err := cli.Do(r, CodecJSON, nil, &failure)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()

	fmt.Printf("%+v\n", failure)
	// Output:
	// {Message:Not Found DocumentationUrl:https://docs.github.com/rest Status:404}
}

func ExampleNewClient_multipartFormData() {
	tf, err := os.CreateTemp(os.TempDir(), "xhttpclient_testdata_*.json")
	if err != nil {
		log.Fatalln(err)
		return
	}
	tfName := tf.Name()
	defer func() {
		if err := os.Remove(tfName); err != nil {
			log.Fatalln(err)
			return
		}
	}()
	if _, err := tf.WriteString(`{"1","file"}`); err != nil {
		log.Fatalln(err)
		return
	}
	if err := tf.Close(); err != nil {
		log.Fatalln(err)
		return
	}

	type Response struct {
		Files struct {
			F string `json:"f"`
			R string `json:"r"`
		} `json:"files"`
		Form struct {
			Hello string `json:"hello"`
		} `json:"form"`
	}

	cli := NewClient()

	r := NewPost().
		Path("https://httpbin.org/post").
		Body([]func(*multipart.Writer) error{
			func(w *multipart.Writer) error {
				return w.WriteField("hello", "world")
			},
			func(w *multipart.Writer) (err error) {
				var pw io.Writer
				if pw, err = w.CreateFormFile("f", filepath.Base(tfName)); err != nil {
					return err
				}

				var file *os.File
				if file, err = os.Open(tfName); err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(pw, file)
				return
			},
			func(w *multipart.Writer) (err error) {
				var pw io.Writer
				if pw, err = w.CreateFormFile("r", "reader.json"); err != nil {
					return err
				}
				_, err = io.Copy(pw, bytes.NewBufferString(`{"2","reader"}`))
				return
			},
		})
	var success Response
	resp, err := cli.Do(r, CodecMultipartFormData, &success, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()

	fmt.Println(success.Form.Hello)
	fmt.Println(success.Files.F)
	fmt.Println(success.Files.R)
	// Output:
	// world
	// {"1","file"}
	// {"2","reader"}
}

func ExampleNewClient_download() {
	cli := NewClient()

	r := NewGet().
		Path("https://raw.githubusercontent.com/electricbubble/xhttpclient/main/LICENSE")
	resp, err := cli.Do(r, CodecNoop, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp.Body())
	if !scanner.Scan() {
		log.Fatalln(scanner.Err())
	}

	fmt.Println(scanner.Text())
	// Output:
	// MIT License
}
