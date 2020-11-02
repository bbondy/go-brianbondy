## Huffman compression

Huffman coding is a mapping of a set of 'symbols' to a set of 'series of bits'. Each symbol corresponds to a variable length series of bits.

The more common symbols are represented with less bits, and the least common symbols are represented with more bits.

There is no algorithm which produces better results than Huffman coding, where each symbol has a binary representation. That is it is, it will always lead to optimal compression.

However, Huffman coding is not the best compression method since it does not take advantage of the relation between symbols.

Example: An 'h' follows a 'T' very often in the english language, and Huffman coding will not take that into consideration.

## Huffman codes

Each Huffman code is prefix free. Meaning no binary code is a prefix of any other binary code. If it was you wouldn't know on the decoding phase which symbol to pick, the prefix of one code which itself a symbol or to keep reading to get another symbol.

## Huffman trees

The binary representations for each symbol is made by using a binary tree. Each symbol is first stored as its own distinct leaf node. Each of these nodes also stores the frequency of the symbol. Since even a single node is considered a tree, each of these nodes are considered their own tree.

From here you simply continuously pick the 2 trees with smallest frequency and combine them into a new tree with a new parent node. The new parent node gets the frequency of both of its children added. This sequence is continued until you have no more trees to merge. The frequency of a tree is defined to be the frequency of its root node.

You now have a single tree with each leaf node being an original symbol. This tree is called a Huffman tree. Creating this tree can be done in O(nlogn) by using a priority queue.

## Getting the code from a Huffman tree

To obtain a code for any symbol you simply start at the root node and follow down the path to the leaf node. For each left arc you follow you add a 0 to the end of your code, for each right arc that you follow you add a 1 to the end of your code.

Since each symbol is a leaf node you are guaranteed that no symbol will be a prefix of any other symbol.

## Combining compression methods

Since Huffman encoding keeps the ordering for symbols to series of bits, other kinds of compression can usually be applied after Huffman coding. For example a compression algorithm that takes into consideration cross boundary symbols will still be useful.

## Adaptive Huffman compression

Adaptive Huffman coding discovers the frequency of symbols as it processes the algorithm.
