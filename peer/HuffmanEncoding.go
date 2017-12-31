// Ported from https://github.com/facebookarchive/RakNet/blob/1a169895a900c9fc4841c556e16514182b75faf8/Source/DS_HuffmanEncodingTree.cpp
package peer
import "container/list"
import "errors"

type huffmanEncodingTreeNode struct {
	value byte
	weight uint32
	left *huffmanEncodingTreeNode
	right *huffmanEncodingTreeNode
	parent *huffmanEncodingTreeNode
}

func insertNodeIntoSortedList(node *huffmanEncodingTreeNode, list *list.List) {
	if list.Len() == 0 {
		list.PushBack(node)
		return
	}
	currentNode := list.Front()

	for true {
		if currentNode.Value.(*huffmanEncodingTreeNode).weight < node.weight {
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

type huffmanEncodingTree struct {
	root *huffmanEncodingTreeNode
	encodingTable [256](struct {
		encoding uint64
		bitLength uint16
	})
}

func GenerateHuffmanFromFrequencyTable(frequencyTable []uint32) *huffmanEncodingTree {
	var counter uint64
	var node *huffmanEncodingTreeNode
	var leafList [256]*huffmanEncodingTreeNode
	huffmanEncodingTreeNodeList := list.New()
	this := &huffmanEncodingTree{}

	for counter = 0; counter < 256; counter++ {
		node = &huffmanEncodingTreeNode{value: byte(counter), weight: frequencyTable[counter], left: nil, right: nil, parent: nil}
		if node.weight == 0 {
			node.weight = 1
		}

		leafList[counter] = node
		insertNodeIntoSortedList(node, huffmanEncodingTreeNodeList)
	}
	
	for true {
		var lesser, greater *huffmanEncodingTreeNode

		lesser = huffmanEncodingTreeNodeList.Remove(huffmanEncodingTreeNodeList.Front()).(*huffmanEncodingTreeNode)
		greater = huffmanEncodingTreeNodeList.Remove(huffmanEncodingTreeNodeList.Front()).(*huffmanEncodingTreeNode)
		node = &huffmanEncodingTreeNode{
			left: lesser,
			right: greater,
			weight: lesser.weight + greater.weight,
		}
		lesser.parent = node
		greater.parent = node

		if huffmanEncodingTreeNodeList.Len() == 0 {
			this.root = node
			this.root.parent = nil
			break
		}

		insertNodeIntoSortedList(node, huffmanEncodingTreeNodeList)
	}

	var tempPath [256]bool
	var tempPathLength uint16
	var currentNode *huffmanEncodingTreeNode

	for counter = 0; counter < 256; counter++ {
		tempPathLength = 0

		currentNode = leafList[counter]
		for do := true; do; do = (currentNode != this.root) {
			if currentNode.parent.left == currentNode {
				tempPath[tempPathLength] = false
				tempPathLength++
			} else {
				tempPath[tempPathLength] = true
				tempPathLength++
			}
			currentNode = currentNode.parent
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

func (tree *huffmanEncodingTree) decodeArray(input *extendedReader, sizeInBits uint, maxCharsToWrite uint, output []byte) error {
	var currentNode *huffmanEncodingTreeNode

	var outputWriteIndex uint = 0
	currentNode = tree.root

	var counter uint
	for counter = 0; counter < sizeInBits; counter++ {
		bit, err := input.readBool()
		if err != nil {
			return err
		}
		if bit == false {
			currentNode = currentNode.left
		} else {
			currentNode = currentNode.right
		}

		if currentNode.left == nil && currentNode.right == nil {
			if outputWriteIndex < maxCharsToWrite {
				output[outputWriteIndex] = currentNode.value
			}
			outputWriteIndex++
			currentNode = tree.root
		}
	}
	return nil
}

func (tree *huffmanEncodingTree) encodeArray(stream *extendedWriter, value []byte) (uint16, error) {
	var err error
	var bitsUsed uint16
	for counter := 0; counter < len(value); counter++ {
		bitsUsed += tree.encodingTable[value[counter]].bitLength
		err = stream.bits(int(tree.encodingTable[value[counter]].bitLength), tree.encodingTable[value[counter]].encoding)
		if err != nil {
			return bitsUsed, err
		}
	}
	if bitsUsed % 8 != 0 {
		remainingBits := 8 - bitsUsed % 8
		for counter := 0; counter < 256; counter++ {
			if tree.encodingTable[counter].bitLength > remainingBits {
				bitsUsed += remainingBits
				return bitsUsed, stream.bits(int(remainingBits), tree.encodingTable[counter].encoding >> (tree.encodingTable[counter].bitLength - remainingBits))
			}
		}
		return bitsUsed, errors.New("could not find encoding longer than remainingBits")
	}
	return bitsUsed, nil
}
