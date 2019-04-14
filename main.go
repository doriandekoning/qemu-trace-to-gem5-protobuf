 package main

import (
	"encoding/binary"
	"bufio"
	"fmt"
	"io"
	"os"
	pb "github.com/doriandekoning/qemu-trace-to-gem5-protobuf/messages"
	"github.com/gogo/protobuf/proto"
)

const ticksPerSec = 1000000000000
const ticksPerNS = ticksPerSec / 1000000000 // 10^9 ns in a sec

var startTimestamp = uint64(0)
var fileSize int64
var progress = int64(0)
var buffer = []byte{}
var bufferIndex uint64
var fileOffset uint64
var inReader *bufio.Reader

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

	inReader = bufio.NewReader(inFile)

	fileSize = getFileSize(inFile)
	fmt.Printf("Total input file size: %dM\n", fileSize/1000000)

	outFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	mappingFilePath := outFilePath + ".mapping"
	mappingFile, err := os.Create(mappingFilePath)
	if err != nil {
	    panic(err)
	}
	defer mappingFile.Close()

	writeFileHeader(outFile)

	readTraceHeader(inFile)
	for true {
		recordType := readUint64(inFile)
		if recordType == 0 {
			readEventMapping(inFile, mappingFile)
		} else if recordType == 1 {
			packet := readTraceEvent(inFile)
			if packet == nil {
			    continue
			}
			marshaledPacket, err := proto.Marshal(packet)
			if err != nil {
				panic(err)
			}
			lengthVarint := proto.EncodeVarint(uint64(len(marshaledPacket)))
			outFile.Write(lengthVarint)
			outFile.Write(marshaledPacket)
		} else {
			panic("Unknown recordType encountered")
		}

		//Get current position
		//TODO refactor to reading new buffer
		offset, err := inFile.Seek(0, 1)
		if err != nil {
			panic(err)
		}
		if (1000 * offset / fileSize) > progress {
			progress = (1000 * offset / fileSize)
			fmt.Printf("Currently %.1f%% done\n", float64(progress)/10)

		}

	}

}

func getFileSize(file *os.File) int64 {
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}

func writeFileHeader(file *os.File) {
	magicnumber := []byte{0x67, 0x65, 0x6d, 0x35}
	_, err := file.Write(magicnumber)
	if err != nil {
		panic(err)
	}

	objectID := "objectid1"
	tickFreq := uint64(1000000000000)
	header := pb.PacketHeader{
		ObjId:    &objectID,
		TickFreq: &tickFreq,
	}
	headerBytes, err := proto.Marshal(&header)
	if err != nil {
		panic(err)
	}

	encodedLength := proto.EncodeVarint(uint64(len(headerBytes)))
	_, err = file.Write(append(encodedLength, headerBytes...))
	if err != nil {
		panic(err)
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

func readTraceEvent(file *os.File) *pb.Packet {
	eventID := readUint64(file)
	if eventID != 75  && eventID != uint64(0xfffffffffffffffe){
		mesg := fmt.Sprintf("Only traces with eventID 75 are supported, found event with id: %d\n", eventID)
		panic(mesg)
	}
	//Read event general data
	timestamp := readUint64(file)
	if startTimestamp == 0 {
		startTimestamp = timestamp
	}
	// RelativeTimestamp is the timestamp in ns from the start of the simulation
	timestampInTicks := ticksPerNS * (timestamp - startTimestamp)
	recLen := readUint32(file)
	readUint32(file) // tracePid	
	if eventID == uint64(0xfffffffffffffffe) {
		readBytes(file, int(recLen))
		fmt.Println("Found dropped trace event")
		return nil
	}
	//Read event arguments (cpu, vaddr and info)
	readUint64(file) // cpu
	vaddr := readUint64(file)
	readUint64(file) //info := USE TO determine cmd
	cmd := uint32(1) //TODO get cmd from info
	//TODO check if size is actually recLen
	return &pb.Packet{Addr: &vaddr, Tick: &timestampInTicks, Size: &recLen, Cmd: &cmd}
}

func readEventMapping(file *os.File, mappingFile *os.File) {
	eventID := readUint64(file)
	length := readUint32(file)
	eventName := string(readBytes(file, int(length)))

	_, err := mappingFile.Write([]byte(fmt.Sprintf( "%d:%s\n", eventID, eventName)))
	if err != nil {
	    panic(err)
	}
}


func readUint64(file *os.File) uint64 {
	return binary.LittleEndian.Uint64(readBytes(file, 8))
}

func readUint32(file *os.File) uint32 {
	//TODO check big or little endian
	return binary.LittleEndian.Uint32(readBytes(file, 4))
}


func readBytes(file *os.File, amount int) []byte {
	if uint64(amount) <= uint64(len(buffer)) - bufferIndex {
		bufferIndex += uint64(amount)
		return buffer[(bufferIndex-uint64(amount)):bufferIndex]
	}
	//Not enough bytes in buffer
	fileOffset += bufferIndex
	if len(buffer) == 0 {
	    buffer = make([]byte, 100000000)
	}
	fmt.Printf("%0.1f%%\n", 100*float64(fileOffset)/float64(fileSize))
	_, err := file.ReadAt(buffer, int64(fileOffset))
	if err == io.EOF {
		fmt.Println("End of file found")
		os.Exit(0)
	} else if err != nil {
		panic(err)
	}
	bufferIndex = uint64(amount)
	return buffer[0:amount]
}
