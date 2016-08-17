// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Generate builtin.go from builtin/runtime.go and builtin/unsafe.go.
// Run this after changing builtin/runtime.go and builtin/unsafe.go
// or after changing the export metadata format in the compiler.
// Either way, you need to have a working compiler binary first.
// See bexport.go for how to make an export metadata format change.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var stdout = flag.Bool("stdout", false, "write to stdout instead of builtin.go")

func main() {
	flag.Parse()

	var b bytes.Buffer
	fmt.Fprintln(&b, "// AUTO-GENERATED by mkbuiltin.go; DO NOT EDIT")
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "package gc")

	mkbuiltin(&b, "runtime")
	mkbuiltin(&b, "unsafe")

	var err error
	if *stdout {
		_, err = os.Stdout.Write(b.Bytes())
	} else {
		err = ioutil.WriteFile("builtin.go", b.Bytes(), 0666)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// Compile .go file, import data from .o file, and write Go string version.
func mkbuiltin(w io.Writer, name string) {
	args := []string{"tool", "compile", "-A"}
	if name == "runtime" {
		args = append(args, "-u")
	}
	args = append(args, "builtin/"+name+".go")

	if err := exec.Command("go", args...).Run(); err != nil {
		log.Fatal(err)
	}
	obj := name + ".o"
	defer os.Remove(obj)

	b, err := ioutil.ReadFile(obj)
	if err != nil {
		log.Fatal(err)
	}

	// Look for $$B that introduces binary export data.
	i := bytes.Index(b, []byte("\n$$B\n"))
	if i < 0 {
		log.Fatal("did not find beginning of export data")
	}
	b = b[i+5:]

	// Look for $$ that closes export data.
	i = bytes.Index(b, []byte("\n$$\n"))
	if i < 0 {
		log.Fatal("did not find end of export data")
	}
	b = b[:i+4]

	// Process and reformat export data.
	const n = 40 // number of bytes per line
	fmt.Fprintf(w, "\nconst %simport = \"\"", name)
	for len(b) > 0 {
		i := len(b)
		if i > n {
			i = n
		}
		fmt.Fprintf(w, " +\n\t%q", b[:i])
		b = b[i:]
	}
	fmt.Fprintf(w, "\n")
}
