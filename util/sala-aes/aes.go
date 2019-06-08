package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"flag"
	"io/ioutil"
	"os"
)

func shuffleSlice(src []byte) []byte {
	ShuffledSrc := make([]byte, 0, len(src))
	ShuffledSrc = append(ShuffledSrc, src[:0x10]...)
	for j := len(src) - 0x10; j >= 0x10; j -= 0x10 {
		ShuffledSrc = append(ShuffledSrc, src[j:j+0x10]...)
	}
	return ShuffledSrc
}

func calculateChecksum(data []byte) uint32 {
	var sum uint32
	var r uint16 = 55665
	var c1 uint16 = 52845
	var c2 uint16 = 22719
	for i := 0; i < len(data); i++ {
		char := data[i]
		cipher := (char ^ byte(r>>8)) & 0xFF
		r = (uint16(cipher)+r)*c1 + c2
		sum += uint32(cipher)
	}
	return sum
}

func main() {
	isDecrypt := flag.Bool("d", false, "If set, will decrypt instead of encrypting")
	outFileName := flag.String("o", "", "Path to output file")
	offset := flag.Int("off", 1, "Offset of data in file")
	flag.Parse()
	inFileName := flag.Arg(0)

	var err error
	var inputFile *os.File

	if inFileName == "" || *outFileName == "" {
		flag.Usage()
		return
	}

	inputFile, err = os.Open(inFileName)
	if err != nil {
		panic(err)
	}

	inputData, err := ioutil.ReadAll(inputFile)
	if err != nil {
		panic(err)
	}
	if *offset >= len(inputData) {
		panic("offset out of bounds")
	}
	inputData = inputData[*offset:]

	// if we are encrypting, write the checksum here
	if !*isDecrypt {
		length := len(inputData)
		paddingSize := 0xF - (length+5)%0x10
		rawCopy := make([]byte, length+6+paddingSize)
		rawCopy[5] = byte(paddingSize & 0xF)
		copy(rawCopy[6+paddingSize:], inputData)

		checkSum := calculateChecksum(rawCopy[4:])
		rawCopy[3] = byte(checkSum >> 24 & 0xFF)
		rawCopy[2] = byte(checkSum >> 16 & 0xFF)
		rawCopy[1] = byte(checkSum >> 8 & 0xFF)
		rawCopy[0] = byte(checkSum & 0xFF)

		inputData = rawCopy
	}

	block, err := aes.NewCipher([]byte{0xFE, 0xF9, 0xF0, 0xEB, 0xE2, 0xDD, 0xD4, 0xCF, 0xC6, 0xC1, 0xB8, 0xB3, 0xAA, 0xA5, 0x9C, 0x97})
	if err != nil {
		panic(err)
	}
	dest := make([]byte, len(inputData))
	c := cipher.NewCBCDecrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	ShuffledSrc := shuffleSlice(inputData)
	c.CryptBlocks(dest, ShuffledSrc)
	dest = shuffleSlice(dest)

	// if we are decrypting, check the checksum here
	if *isDecrypt {
		checkSum := calculateChecksum(dest[4:])
		storedChecksum := binary.LittleEndian.Uint32(dest[:4])
		if storedChecksum != checkSum {
			println("warning: checksum fail")
		}
		paddingSize := dest[6] & 0xF
		dest = dest[7+paddingSize:]
	}

	err = ioutil.WriteFile(*outFileName, dest, 0666)
	if err != nil {
		panic(err)
	}
}
