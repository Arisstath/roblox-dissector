// Ported from https://github.com/facebookarchive/RakNet/blob/1a169895a900c9fc4841c556e16514182b75faf8/Source/DS_HuffmanEncodingTree.cpp
package peer
import "container/list"
import "errors"

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

	for true {
		if currentNode.Value.(*HuffmanEncodingTreeNode).Weight < node.Weight {
			currentNode = currentNode.Next()
			if currentNode == nil { // end reached
				list.PushBack(node)
				break
			}
		} else {
			list.InsertBefore(node, currentNode)
			break
		}
	}
}

type HuffmanEncodingTree struct {
	root *HuffmanEncodingTreeNode
	encodingTable [256](struct {
		encoding uint64
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

		var buffer uint64
		bitlen := tempPathLength
		for tempPathLength > 0 {
			tempPathLength--
			buffer <<= 1
			if tempPath[tempPathLength] {
				buffer |= 1
			}
		}
		this.encodingTable[counter].bitLength = bitlen
		this.encodingTable[counter].encoding = buffer
	}
	return this
}

func (tree *HuffmanEncodingTree) DecodeArray(input *ExtendedReader, sizeInBits uint, maxCharsToWrite uint, output []byte) error {
	var currentNode *HuffmanEncodingTreeNode

	var outputWriteIndex uint = 0
	currentNode = tree.root

	var counter uint
	for counter = 0; counter < sizeInBits; counter++ {
		bit, err := input.ReadBool()
		if err != nil {
			return err
		}
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
	return nil
}

func (tree *HuffmanEncodingTree) EncodeArray(stream *ExtendedWriter, value []byte) (uint16, error) {
	var err error
	var bitsUsed uint16
	for counter := 0; counter < len(value); counter++ {
		bitsUsed += tree.encodingTable[value[counter]].bitLength
		err = stream.Bits(int(tree.encodingTable[value[counter]].bitLength), tree.encodingTable[value[counter]].encoding)
		if err != nil {
			return bitsUsed, err
		}
	}
	if bitsUsed % 8 != 0 {
		remainingBits := 8 - bitsUsed % 8
		for counter := 0; counter < 256; counter++ {
			if tree.encodingTable[counter].bitLength > remainingBits {
				bitsUsed += remainingBits
				return bitsUsed, stream.Bits(int(remainingBits), tree.encodingTable[counter].encoding >> (tree.encodingTable[counter].bitLength - remainingBits))
			}
		}
		return bitsUsed, errors.New("could not find encoding longer than remainingBits")
	}
	return bitsUsed, nil
}
