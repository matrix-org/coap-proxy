package gson

type bufferhead struct {
	head *buffernode
}

func (b *bufferhead) getbuffer(size int) *buffernode {
	if node := b.head; node != nil {
		node.data = fixbuffer(node.data, int64(size))
		b.head = node.next
		return node
	}
	node := &buffernode{data: fixbuffer(nil, int64(size)), next: nil}
	return node
}

func (b *bufferhead) putbuffer(node *buffernode) {
	node.next = b.head
	b.head = node
}

type buffernode struct {
	data []byte
	next *buffernode
}
