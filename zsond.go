package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func parsePath(path string) (string, error) {
	dirpath := filepath.Join(".", path)
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return "", err
	}
	return dirpath, nil
}

func findPath(buf []byte) string {
	for len(buf) > 5 {
		off := bytes.IndexByte(buf, byte('\n'))
		if off < 0 {
			return ""
		}
		line := string(buf[:off])
		if strings.HasPrefix(line, "#path") {
			return strings.TrimSpace(line[5:])
		}
		buf = buf[off+1:]
	}
	return ""
}

func openfile(path, name string) (*os.File, string, error) {
	for n := 0; n < 100; n++ {
		var fname string
		if n > 0 {
			fname = fmt.Sprintf("%s.%d.log", name, n)
		} else {
			fname = fmt.Sprintf("%s.log", name)
		}
		filename := filepath.Join(path, fname)
		f, err := os.OpenFile(filename, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			return nil, "", err
		}
		return f, filename, nil
	}
	return nil, "", errors.New("too many files")
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad method", http.StatusForbidden)
		return
	}
	dirPath, err := parsePath(r.URL.RequestURI())
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	size := 128 * 1024
	reader := bufio.NewReaderSize(r.Body, size)
	buf, err := reader.Peek(size)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if len(buf) == 0 {
		http.Error(w, "empty body", http.StatusForbidden)
		return
	}
	logType := findPath(buf)
	if logType == "" {
		http.Error(w, "no #path directive found", http.StatusForbidden)
		return
	}
	file, name, err := openfile(dirPath, logType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	go func() {
		for range time.Tick(time.Second) {
			file.Sync()
		}
	}()
	if _, err := io.Copy(file, reader); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("wrote %s\n", name)
}

func main() {
	port := ":9867"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
