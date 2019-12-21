package main

import (
	"fmt"
	"github.com/pin/tftp"
	"io"
	"os"
	"time"
)

func readHandler(filename string, rf io.ReaderFrom) error {
	go func() {
		file, err := os.Open("./fileserver" + filename)
		defer file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		n, err := rf.ReadFrom(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		fmt.Printf("%d bytes sent\n", n)
	}()
	return nil
}

func writeHandler(filename string, wt io.WriterTo) error {
	return nil
}
func main() {
	// use nil in place of handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetTimeout(5 * time.Second)    // optional
	err := s.ListenAndServe(":4444") // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}
