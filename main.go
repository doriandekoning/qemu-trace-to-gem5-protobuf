package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func main() {
	inFilePath := os.Args[1]
	if inFilePath == "" {
		fmt.Println("No input file provided")
	}
	fmt.Println("Using input file: ", os.Args[1])

	outFilePath := os.Args[2]
	if outFilePath == "" {
		fmt.Println("No output file is provided")
	}
	fmt.Println("Using output file:", os.Args[2])

	inFile, err := os.Open(inFilePath)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	readTraceHeader(inFile)

	for true {
		recordType := readUint64(inFile)
		if recordType == 0 {
			readEventMapping(inFile)
		} else if recordType == 1 {
			readTraceEvent(inFile)
		} else {
			panic("Unknown recordType encountered")
		}
	}

}

func readTraceHeader(file *os.File) {
	eventID := readUint64(file)
	fmt.Println("EventID: ", eventID)
	//Nothing to check here
	magicNumber := readUint64(file)
	if magicNumber != uint64(0xf2b177cb0aa429b4) {
		panic("Wrong magic number encountered")
	}
	headerVersion := readUint64(file)
	if headerVersion != 4 {
		panic("Only header version 4 is supported")
	}
}

func readTraceEvent(file *os.File) {
	eventID := readUint64(file)
	if eventID != 75 {
		panic("Only traces with only event 75 are supported")
	}
	//Read event general data
	timestamp := readUint64(file)
	recLen := readUint32(file)
	tracePid := readUint32(file)
	//Read event arguments (cpu, vaddr and info)
	cpu := readUint64(file)
	vaddr := readUint64(file)
	info := readUint64(file)

	fmt.Printf("Read event timestamp: %d pid:%d cpu:%x vaddr:%08X info:%x %d\n", timestamp, tracePid, cpu, vaddr, info, recLen)

}

func readEventMapping(file *os.File) {
	eventID := readUint64(file)
	length := readUint32(file)
	name := readBytes(file, int(length))
	fmt.Println(eventID, ":", string(name))
}

func readUint64(file *os.File) uint64 {
	return binary.LittleEndian.Uint64(readBytes(file, 8))
}

func readUint32(file *os.File) uint32 {
	//TODO check big or little endian
	return binary.LittleEndian.Uint32(readBytes(file, 4))
}

func finish(file *os.File) {
	file.Close()
	os.Exit(0)
}

func readBytes(file *os.File, amount int) []byte {
	bytes := make([]byte, amount)

	_, err := file.Read(bytes)
	if err == io.EOF {
		fmt.Println("End of file found")
		finish(file)
	} else if err != nil {
		panic(err)
	}
	return bytes
}
