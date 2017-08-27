// Ported from https://github.com/facebookarchive/RakNet/blob/1a169895a900c9fc4841c556e16514182b75faf8/Source/DS_HuffmanEncodingTree.cpp
package peer
import "container/list"
import "github.com/gskartwii/go-bitstream"
import "bytes"

type HuffmanEncodingTreeNode struct {
	Value byte
	Weight uint32
	Left *HuffmanEncodingTreeNode
	Right *HuffmanEncodingTreeNode
	Parent *HuffmanEncodingTreeNode
}

func InsertNodeIntoSortedList(node *HuffmanEncodingTreeNode, list *list.List) {
	if list.Len() == 0 {
		list.PushBack(node)
		return
	}
	currentNode := list.Front()
	var counter int

	for true {
		nextNode := currentNode.Next()
		if nextNode == nil { // end reached
			list.PushBack(node)
			break
		}
		if nextNode.Value.(*HuffmanEncodingTreeNode).Weight < node.Weight {
			currentNode = nextNode
		} else {
			list.InsertAfter(node, currentNode)
			break
		}

		counter++
		if counter == list.Len() {
			list.PushBack(node)
			break
		}
	}
}

type HuffmanEncodingTree struct {
	root *HuffmanEncodingTreeNode
	encodingTable [256](struct {
		encoding []byte
		bitLength uint16
	})
}

func GenerateHuffmanFromFrequencyTable(frequencyTable []uint32) *HuffmanEncodingTree {
	var counter uint64
	var node *HuffmanEncodingTreeNode
	var leafList [256]*HuffmanEncodingTreeNode
	huffmanEncodingTreeNodeList := list.New()
	this := &HuffmanEncodingTree{}

	for counter = 0; counter < 256; counter++ {
		node = &HuffmanEncodingTreeNode{Value: byte(counter), Weight: frequencyTable[counter], Left: nil, Right: nil, Parent: nil}
		if node.Weight == 0 {
			node.Weight = 1
		}

		leafList[counter] = node
		InsertNodeIntoSortedList(node, huffmanEncodingTreeNodeList)
	}
	
	for true {
		var lesser, greater *HuffmanEncodingTreeNode

		lesser = huffmanEncodingTreeNodeList.Remove(huffmanEncodingTreeNodeList.Front()).(*HuffmanEncodingTreeNode)
		greater = huffmanEncodingTreeNodeList.Remove(huffmanEncodingTreeNodeList.Front()).(*HuffmanEncodingTreeNode)
		node = &HuffmanEncodingTreeNode{
			Left: lesser,
			Right: greater,
			Weight: lesser.Weight + greater.Weight,
		}
		lesser.Parent = node
		greater.Parent = node

		if huffmanEncodingTreeNodeList.Len() == 0 {
			this.root = node
			this.root.Parent = nil
			break
		}

		InsertNodeIntoSortedList(node, huffmanEncodingTreeNodeList)
	}

	var tempPath [256]bool
	var tempPathLength uint16
	var currentNode *HuffmanEncodingTreeNode

	for counter = 0; counter < 256; counter++ {
		tempPathLength = 0

		currentNode = leafList[counter]
		for do := true; do; do = (currentNode != this.root) {
			if currentNode.Parent.Left == currentNode {
				tempPath[tempPathLength] = false
				tempPathLength++
			} else {
				tempPath[tempPathLength] = true
				tempPathLength++
			}
			currentNode = currentNode.Parent
		}

		buffer := make([]byte, (tempPathLength + 7)/8)
		bitStream := bitstream.NewWriter(bytes.NewBuffer(buffer))
		bitlen := tempPathLength
		for tempPathLength > 0 {
			tempPathLength--
			bitStream.WriteBit(bitstream.Bit(tempPath[tempPathLength]))
		}
		bitStream.Flush(bitstream.Bit(false))
		this.encodingTable[counter].bitLength = bitlen
		this.encodingTable[counter].encoding = buffer
	}
	return this
}

func (tree *HuffmanEncodingTree) DecodeArray(input *ExtendedReader, sizeInBits uint, maxCharsToWrite uint, output []byte) {
	var currentNode *HuffmanEncodingTreeNode

	var outputWriteIndex uint = 0
	currentNode = tree.root

	var counter uint
	for counter = 0; counter < sizeInBits; counter++ {
		bit, _ := input.ReadBool()
		if bit == false {
			currentNode = currentNode.Left
		} else {
			currentNode = currentNode.Right
		}

		if currentNode.Left == nil && currentNode.Right == nil {
			if outputWriteIndex < maxCharsToWrite {
				output[outputWriteIndex] = currentNode.Value
			}
			outputWriteIndex++
			currentNode = tree.root
		}
	}
}
