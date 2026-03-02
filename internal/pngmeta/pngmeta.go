package pngmeta

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

var pngSignature = []byte{137, 80, 78, 71, 13, 10, 26, 10}

// compressionThreshold is the size above which we compress the text in iTXt chunks.
const compressionThreshold = 1024

// EmbedText inserts an iTXt chunk with the given keyword and text into pngData.
// The chunk is inserted after the IHDR chunk (first chunk), before any other chunks.
func EmbedText(pngData []byte, keyword, text string) ([]byte, error) {
	if len(pngData) < len(pngSignature) || !bytes.Equal(pngData[:len(pngSignature)], pngSignature) {
		return nil, fmt.Errorf("not a valid PNG file")
	}

	// Find the end of the IHDR chunk (signature + IHDR chunk)
	pos := len(pngSignature)
	if pos+8 > len(pngData) {
		return nil, fmt.Errorf("PNG too short to contain IHDR")
	}

	ihdrLen := binary.BigEndian.Uint32(pngData[pos : pos+4])
	chunkType := string(pngData[pos+4 : pos+8])
	if chunkType != "IHDR" {
		return nil, fmt.Errorf("first chunk is not IHDR")
	}

	// IHDR chunk: 4 (length) + 4 (type) + ihdrLen (data) + 4 (CRC)
	ihdrEnd := pos + 4 + 4 + int(ihdrLen) + 4

	// Build the iTXt chunk data
	itxtData := buildITXtData(keyword, text)

	// Build the full chunk: length + type + data + CRC
	chunk := buildChunk("iTXt", itxtData)

	// Insert the chunk after IHDR
	result := make([]byte, 0, len(pngData)+len(chunk))
	result = append(result, pngData[:ihdrEnd]...)
	result = append(result, chunk...)
	result = append(result, pngData[ihdrEnd:]...)

	return result, nil
}

// ExtractText finds an iTXt chunk with the given keyword and returns its text content.
func ExtractText(pngData []byte, keyword string) (string, error) {
	if len(pngData) < len(pngSignature) || !bytes.Equal(pngData[:len(pngSignature)], pngSignature) {
		return "", fmt.Errorf("not a valid PNG file")
	}

	pos := len(pngSignature)
	keywordBytes := []byte(keyword)

	for pos+8 <= len(pngData) {
		chunkLen := binary.BigEndian.Uint32(pngData[pos : pos+4])
		chunkType := string(pngData[pos+4 : pos+8])
		dataStart := pos + 8
		dataEnd := dataStart + int(chunkLen)

		if dataEnd+4 > len(pngData) {
			return "", fmt.Errorf("truncated PNG chunk")
		}

		if chunkType == "iTXt" {
			text, ok := parseITXtData(pngData[dataStart:dataEnd], keywordBytes)
			if ok {
				return text, nil
			}
		}

		if chunkType == "IEND" {
			break
		}

		// Move to next chunk: 4 (length) + 4 (type) + chunkLen (data) + 4 (CRC)
		pos = dataEnd + 4
	}

	return "", fmt.Errorf("no iTXt chunk with keyword %q found", keyword)
}

// buildITXtData constructs the data portion of an iTXt chunk.
// Format: keyword\0 compressionFlag compressionMethod languageTag\0 translatedKeyword\0 text
func buildITXtData(keyword, text string) []byte {
	var buf bytes.Buffer
	buf.WriteString(keyword)
	buf.WriteByte(0) // null separator after keyword

	if len(text) > compressionThreshold {
		// Compressed
		buf.WriteByte(1) // compression flag
		buf.WriteByte(0) // compression method (zlib)
		buf.WriteByte(0) // language tag (empty, null-terminated)
		buf.WriteByte(0) // translated keyword (empty, null-terminated)

		var compressed bytes.Buffer
		w := zlib.NewWriter(&compressed)
		w.Write([]byte(text))
		w.Close()
		buf.Write(compressed.Bytes())
	} else {
		// Uncompressed
		buf.WriteByte(0) // compression flag
		buf.WriteByte(0) // compression method
		buf.WriteByte(0) // language tag (empty, null-terminated)
		buf.WriteByte(0) // translated keyword (empty, null-terminated)
		buf.WriteString(text)
	}

	return buf.Bytes()
}

// parseITXtData parses an iTXt chunk data and returns the text if keyword matches.
func parseITXtData(data, keyword []byte) (string, bool) {
	// Find keyword null terminator
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx < 0 {
		return "", false
	}

	if !bytes.Equal(data[:nullIdx], keyword) {
		return "", false
	}

	// After keyword\0: compressionFlag(1) compressionMethod(1) languageTag\0 translatedKeyword\0 text
	rest := data[nullIdx+1:]
	if len(rest) < 2 {
		return "", false
	}

	compressionFlag := rest[0]
	// compressionMethod := rest[1]
	rest = rest[2:]

	// Skip language tag (null-terminated)
	nullIdx = bytes.IndexByte(rest, 0)
	if nullIdx < 0 {
		return "", false
	}
	rest = rest[nullIdx+1:]

	// Skip translated keyword (null-terminated)
	nullIdx = bytes.IndexByte(rest, 0)
	if nullIdx < 0 {
		return "", false
	}
	rest = rest[nullIdx+1:]

	if compressionFlag == 1 {
		// zlib decompress
		r, err := zlib.NewReader(bytes.NewReader(rest))
		if err != nil {
			return "", false
		}
		defer r.Close()
		decompressed, err := io.ReadAll(r)
		if err != nil {
			return "", false
		}
		return string(decompressed), true
	}

	return string(rest), true
}

// buildChunk builds a full PNG chunk with length, type, data, and CRC.
func buildChunk(chunkType string, data []byte) []byte {
	buf := make([]byte, 4+4+len(data)+4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(data)))
	copy(buf[4:8], chunkType)
	copy(buf[8:], data)

	crc := crc32.NewIEEE()
	crc.Write(buf[4 : 8+len(data)])
	binary.BigEndian.PutUint32(buf[8+len(data):], crc.Sum32())

	return buf
}
