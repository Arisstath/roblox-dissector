package peer

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestClusterSerialize(t *testing.T) {
	packet, err := ioutil.ReadFile("./testpackets/cluster.bin")
	if err != nil {
		t.Fatal("reading cluster:", err.Error())
	}
	reader := &extendedReader{bytes.NewReader(packet)}
	chunks, err := deserializeChunks(reader)
	if err != nil {
		t.Fatal("parsing cluster:", err.Error())
	}

	writeBuf := bytes.NewBuffer(nil)
	writer := &extendedWriter{writeBuf}
	err = (&Packet8DLayer{Chunks: chunks}).serializeChunks(writer)
	if err != nil {
		t.Fatal("writing cluster:", err.Error())
	}

	result := bytes.Equal(packet, writeBuf.Bytes())
	if !result {
		ioutil.WriteFile("./testpackets/cluster.out", writeBuf.Bytes(), 0666)
		t.Error("bytes were inequal, see testpackets/cluster.out")
	}
}
