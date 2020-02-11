package main

import (
	"archive/zip"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/godfried/osmimport/sources/trig"
)

func main() {
	log.SetOutput(os.Stdout)
	trigSource := flag.String("dir", "", "path to KML directory with Trig data")
	out := flag.String("out", "trig-poi-all.gob", "path to output file")
	flag.Parse()
	files, err := filepath.Glob(*trigSource + "/*.kmz")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	output, err := os.Create(*out)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer output.Close()
	e := gob.NewEncoder(output)

	for _, f := range files {
		err = extractFiles(f, e)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func extractFiles(in string, output *gob.Encoder) error {
	zr, err := zip.OpenReader(in)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		err := extractFile(f, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractFile(f *zip.File, output *gob.Encoder) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	trigs, err := trig.Read(rc)
	if err != nil {
		return err
	}
	for _, t := range trigs {
		err = output.Encode(t)
		if err != nil {
			return err
		}
	}
	return nil
}
