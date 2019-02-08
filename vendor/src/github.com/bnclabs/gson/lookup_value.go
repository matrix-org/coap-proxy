package gson

import "strconv"
import "fmt"

func valGet(segments [][]byte, doc interface{}) interface{} {
	if len(segments) == 0 { // exit recursion.
		return doc
	}

	segment := bytes2str(segments[0])

	switch val := doc.(type) {
	case []interface{}:
		if segment == "-" { // not req. as per rfc-6901
			return valGet(segments[1:], val[len(val)-1])
		} else if idx, err := strconv.Atoi(segment); err != nil {
			panic("valGet(): gson pointer-invalidIndex")
		} else if idx >= len(val) {
			panic("valGet(): gson pointer-index-outofRange")
		} else {
			return valGet(segments[1:], val[idx])
		}

	case map[string]interface{}:
		if doc, ok := val[segment]; !ok {
			panic(fmt.Sprintf("valGet(): gson %v pointer-invalidKey", segment))
		} else {
			return valGet(segments[1:], doc)
		}
	}
	panic("valGet(): gson invalidPointer")
}

func valSet(segments [][]byte, doc, item interface{}) (newdoc, old interface{}) {
	ln, container := len(segments), doc
	if ln == 0 {
		panic("valSet(): document is not a container")
	} else if ln > 1 {
		container = valGet(segments[:ln-1], doc)
	} // else if ln == 1, container _is_ doc

	var ok bool
	switch cont := container.(type) {
	case []interface{}:
		key := bytes2str(segments[ln-1])
		if key == "-" {
			old, cont[len(cont)-1] = cont[len(cont)-1], item
			if ln > 1 {
				doc, _ = valSet(segments[:ln-1], doc, cont)
			} else { // edge case !
				doc = cont
			}
		} else if idx, err := strconv.Atoi(key); err != nil {
			panic("valSet(): gson pointer-invalidIndex")
		} else if idx >= len(cont) {
			panic("valSet(): gson pointer-outofRange")
		} else {
			old, cont[idx] = cont[idx], item
		}

	case map[string]interface{}:
		key := string(segments[ln-1])
		if old, ok = cont[key]; !ok {
			old = item
		}
		cont[key] = item
	default:
		panic("valSet(): gson invalidPointer")
	}
	return doc, old
}

func valDel(segments [][]byte, doc interface{}) (newdoc, old interface{}) {
	ln, container := len(segments), doc
	if ln == 0 {
		panic("valDel(): document is not a container")
	} else if ln > 1 {
		container = valGet(segments[:ln-1], doc)
	} // else if ln == 1, container _is_ doc

	key := bytes2str(segments[ln-1])

	switch cont := container.(type) {
	case []interface{}:
		if idx, err := strconv.Atoi(key); err != nil {
			fmsg := fmt.Errorf("valDel(): gson pointer-invalidIndex `%v`", err)
			panic(fmsg)
		} else if idx >= len(cont) {
			panic("valDel(): gson pointer-outofRange")
		} else {
			old = cont[idx]
			copy(cont[idx:], cont[idx+1:])
			cont = cont[:len(cont)-1]
			if ln > 1 {
				doc, _ = valSet(segments[:ln-1], doc, cont)
				return doc, old
			}
			// edge case !!
			return cont, old
		}

	case map[string]interface{}:
		old, _ = cont[key]
		delete(cont, key)

	default:
		panic("valDel(): gson invalidPointer")
	}
	return doc, old
}

func valAppend(segments [][]byte, doc, item interface{}) (newdoc interface{}) {
	container := doc
	if len(segments) > 0 {
		container = valGet(segments, doc)
	} // else if ln == 1, container _is_ doc

	switch cont := container.(type) {
	case []interface{}:
		cont = append(cont, item)
		if len(segments) == 0 {
			newdoc = cont
		} else {
			newdoc, _ = valSet(segments, doc, cont)
		}

	default:
		panic("valAppend(): invalidPointer")
	}
	return newdoc
}

func valPrepend(segments [][]byte, doc, item interface{}) (newdoc interface{}) {
	container := doc
	if len(segments) > 0 {
		container = valGet(segments, doc)
	} // else if ln == 1, container _is_ doc

	switch cont := container.(type) {
	case []interface{}:
		ln := len(cont)
		cont = append(cont, nil)
		copy(cont[1:], cont[:ln])
		cont[0] = item
		if len(segments) == 0 {
			newdoc = cont
		} else {
			newdoc, _ = valSet(segments, doc, cont)
		}

	default:
		panic("valPrepend(): invalidPointer")
	}
	return newdoc
}
