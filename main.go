package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

type root struct {
	Files map[string]fileEntry `json:"files"`
}

type fileEntry struct {
	Size   int64  `json:"size"`
	Offset string `json:"offset"`
}

func main() {
	cwd, _ := os.Getwd()

	inputPath := flag.String("i", cwd, "Input path")
	outPath := flag.String("o", path.Join(cwd, "out.asar"), "Output path")

	flag.Parse()

	fileEntries := make(map[string]fileEntry)

	files, _ := ioutil.ReadDir(*inputPath)

	var offset int64

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fileEntries[f.Name()] = fileEntry{
			Size:   f.Size(),
			Offset: strconv.FormatInt(offset, 10),
		}
		offset += f.Size()
	}

	header, err := json.Marshal(root{Files: fileEntries})

	if err != nil {
		log.Fatal("Marshal:", err)
	}

	out, err := os.Create(*outPath)
	if err != nil {
		log.Fatal("file open error:", err)
	}
	defer out.Close()

	headerPadding := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerPadding, uint32(len(header)))

	headerPaddingPadding := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerPaddingPadding, uint32(len(header)+len(headerPadding)))

	headerSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSize, uint32(len(header)+len(headerPadding)+len(headerPaddingPadding)))

	headerSizePadding := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSizePadding, uint32(len(headerSize)))

	out.Write(headerSizePadding)
	out.Write(headerSize)
	out.Write(headerPaddingPadding)
	out.Write(headerPadding)
	out.Write(header)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		content, err := ioutil.ReadFile(path.Join(*inputPath, f.Name()))
		if err != nil {
			log.Fatal("file read error", err)
		}
		out.Write(content)
	}
}
