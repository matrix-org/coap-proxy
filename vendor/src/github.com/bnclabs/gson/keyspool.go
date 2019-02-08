package gson

type mkeyshead struct {
	head *mkeysnode
}

func (b *mkeyshead) getmkeys(size int) *mkeysnode {
	if node := b.head; node != nil {
		if len(node.keys) < size {
			node.keys = make([]string, 0, size)
		}
		node.keys = node.keys[:0]
		b.head = node.next
		return node
	}
	node := &mkeysnode{keys: make([]string, 0, size), next: nil}
	return node
}

func (b *mkeyshead) putmkeys(node *mkeysnode) {
	node.next = b.head
	b.head = node
}

type mkeysnode struct {
	keys []string
	next *mkeysnode
}

// sort JSON property objects based on property names.

func (m *mkeysnode) sortProps1(val map[string]interface{}) []string {
	for k := range val {
		m.keys = append(m.keys, k)
	}

	m.keys = sortStrings(m.keys)
	return m.keys
}

func (m *mkeysnode) sortProps2(val map[string]uint64) []string {
	for k := range val {
		m.keys = append(m.keys, k)
	}
	m.keys = sortStrings(m.keys)
	return m.keys
}

func (m *mkeysnode) sortProps3(val [][2]interface{}) []string {
	for _, item := range val {
		m.keys = append(m.keys, item[0].(string))
	}
	m.keys = sortStrings(m.keys)
	return m.keys
}
