package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Ohai!")

	//processzip.Testing()

	data := []byte("The quick brown fox jumped over the lazy dug.")

	pos := int64(0)
	length := int64(44)

	size := int64(len(data))
	chunkCount := size / length
	fmt.Printf("Chunk count: %d / %d = %d\n", size, length, chunkCount)
	if size % length > 0 {
		fmt.Printf("Bumping chunk count by one for remainder (%d)\n", size % length)
		chunkCount++
	}
	chunks := make([][]byte, chunkCount + 1)
	fmt.Printf("Downloading %d chunks of length %d from data of size %d\n", chunkCount, length, size)

	index := 0

	var wg sync.WaitGroup
	for {
		if pos < size {
			wg.Add(1)
			if pos + length < size {
				fmt.Printf("Downloading %d bytes from position %d to %d (total %d) - chunk %d\n", length, pos, (pos+length) - 1, len(data), index)
				go downloadChunk(data, index, chunks, pos, length, &wg)
			} else {
				fmt.Printf("Downloading the remainder from position %d (total %d) - chunk %d\n", pos, len(data), index)
				go downloadChunk(data, index, chunks, pos, -1, &wg)
			}
			index++
			pos += length
		} else {
			break
		}
	}
	wg.Wait()

	result := assemble(chunks)

	fmt.Printf("data: %v\n", string(data))
	fmt.Printf("data: %v\n", string(result))
}

func downloadChunk(data []byte, index int, chunks [][]byte, pos int64, length int64, wg *sync.WaitGroup) {
	defer wg.Done()

	var read []byte
	if length > 0 {
		read = data[pos:pos + length]
	} else {
		read = data[pos:]
	}
	chunks[index] = read
}

func assemble(chunks [][]byte) []byte {
	var result []byte
	for _, chunk := range chunks {
		result = append(result, chunk...)
	}
	return result
}