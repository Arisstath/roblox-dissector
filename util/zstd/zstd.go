package main

import (
	"flag"
	"io"
	"os"

	"github.com/DataDog/zstd"
)

func main() {
	isDecompress := flag.Bool("d", false, "If set, will decompress instead of compress")
	outFileName := flag.String("o", "", "Path to output file")
	offset := flag.Int("off", 1, "Offset of data in file")
	flag.Parse()
	inFileName := flag.Arg(0)

	var err error
	var inputFile *os.File
	var inputStream io.ReadCloser
	var outputFile *os.File
	var outputStream io.WriteCloser

	if inFileName == "" || *outFileName == "" {
		flag.Usage()
		return
	}

	inputFile, err = os.Open(inFileName)
	if err != nil {
		panic(err)
	}
	_, err = inputFile.Seek(int64(*offset), io.SeekStart)
	if err != nil {
		panic(err)
	}
	outputFile, err = os.OpenFile(*outFileName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}

	if *isDecompress {
		println("Will decompress")
		inputStream = zstd.NewReader(inputFile)
		outputStream = outputFile
	} else {
		println("Will compress")
		inputStream = inputFile
		outputStream = zstd.NewWriter(outputFile)
	}

	_, err = io.Copy(outputStream, inputStream)
	if err != nil {
		panic(err)
	}

	err = inputFile.Close()
	if err != nil {
		panic(err)
	}
	err = outputFile.Close()
	if err != nil {
		panic(err)
	}

	println("Successfully completed operation")
}
