package gson

type kvhead struct {
	head *kvnode
}

func (b *kvhead) getkv(size int) *kvnode {
	if node := b.head; node != nil {
		if cap(node.refs) < size {
			node.refs = make(kvrefs, 0, size)
		}
		node.refs = node.refs[:0]
		b.head = node.next
		return node
	}
	node := &kvnode{refs: make(kvrefs, 0, size), next: nil}
	return node
}

func (b *kvhead) putkv(node *kvnode) {
	node.next = b.head
	b.head = node
}

type kvnode struct {
	refs kvrefs
	next *kvnode
}

//---- data modelling to sort and collate JSON property items.

type kvref struct {
	key  string
	code []byte
}

type kvrefs []kvref

func (kv kvrefs) Len() int {
	return len(kv)
}

func (kv kvrefs) Less(i, j int) bool {
	return kv[i].key < kv[j].key
}

func (kv kvrefs) Swap(i, j int) {
	tmp := kv[i]
	kv[i] = kv[j]
	kv[j] = tmp
}

// bubble sort, moving to qsort should be atleast 40% faster.
func (kv kvrefs) sort() {
	for ln := len(kv) - 1; ; ln-- {
		changed := false
		for i := 0; i < ln; i++ {
			if kv[i].key > kv[i+1].key {
				kv[i], kv[i+1] = kv[i+1], kv[i]
				changed = true
			}
		}
		if changed == false {
			break
		}
	}
}
