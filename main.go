package main

import (
	"encoding/binary"
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

	fileSize = getFileSize(inFile)
	fmt.Printf("Total input file size: %dM\n", fileSize/1000000)

	outFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	writeFileHeader(outFile)

	readTraceHeader(inFile)
	for true {
		recordType := readUint64(inFile)
		if recordType == 0 {
			readEventMapping(inFile, false)
		} else if recordType == 1 {
			packet := readTraceEvent(inFile)
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
		offset, err := inFile.Seek(0, 1)
		if err != nil {
			panic(err)
		}
		if (100 * offset / fileSize) > progress {
			progress = (100 * offset / fileSize)
			fmt.Printf("Currently %d%% done\n", progress)

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
	if eventID != 75 {
		panic("Only traces with only event 75 are supported")
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
	//Read event arguments (cpu, vaddr and info)
	readUint64(file) // cpu
	vaddr := readUint64(file)
	readUint64(file) //info := USE TO determine cmd
	cmd := uint32(1) //TODO get cmd from info
	//TODO check if size is actually recLen
	return &pb.Packet{Addr: &vaddr, Tick: &timestampInTicks, Size: &recLen, Cmd: &cmd}
}

func readEventMapping(file *os.File, print bool) {
	eventID := readUint64(file)
	length := readUint32(file)
	eventName := readBytes(file, int(length))
	if print {
		fmt.Println(eventID, ":", eventName)
	}
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
