// implements rfc-6901 to parse json-pointer text and
// work with golang arrays and maps

package gson

import "strconv"
import "strings"
import "unicode/utf8"

// MaxJsonpointerLen size of json-pointer path. Affects memory pool.
// Changing this value will affect all new configuration instances.
var MaxJsonpointerLen = 2048

type jptrConfig struct {
	jptrMaxlen int
	jptrMaxseg int
}

// SetJptrlen will set the maximum size for jsonpointer path.
func (config Config) SetJptrlen(n int) *Config {
	config.jptrMaxlen = n
	config.jptrMaxseg = n / 8
	return &config
}

// Jsonpointer abstracts rfc-6901 into a type.
// allows ~0 and ~1 escapes, property lookup by specifying the key,
// and array lookup by specifying the index.
// Also allows empty "" pointer and empty key "/".
type Jsonpointer struct {
	config   *Config
	path     []byte
	segments [][]byte
	segln    int
}

// ResetPath to reuse the Jsonpointer object for a new path.
func (jptr *Jsonpointer) ResetPath(path string) *Jsonpointer {
	if len(path) > jptr.config.jptrMaxlen {
		panic("jsonpointer path exceeds configured length")
	}
	n := copy(jptr.path[:cap(jptr.path)], path)
	jptr.path = jptr.path[:n]
	jptr.segln = 0
	return jptr
}

// ResetSegments variant of ResetPath to reconstruct the path from segments.
func (jptr *Jsonpointer) ResetSegments(segments []string) *Jsonpointer {
	if len(segments) > jptr.config.jptrMaxseg {
		panic("no. of segments in jsonpointer-path exceeds configured limit")
	}
	n := encodePointer(segments, jptr.path[:cap(jptr.path)])
	jptr.path = jptr.path[:n]
	for i, segment := range segments {
		jptr.segments[i] = append(jptr.segments[i][:0], segment...)
	}
	jptr.segln = len(segments)
	return jptr
}

// Segments return path segments, segments in a path are separated by "/"
func (jptr *Jsonpointer) Segments() [][]byte {
	if len(jptr.path) > 0 && jptr.segln == 0 {
		jptr.segln = parsePointer(jptr.path, jptr.segments)
	}
	return jptr.segments[:jptr.segln]
}

// Path return the path value.
func (jptr *Jsonpointer) Path() []byte {
	return jptr.path
}

//---- local functions

func parsePointer(in []byte, segments [][]byte) int {
	j := 0

	if len(in) == 0 {
		return 0
	} else if len(in) == 1 && in[0] == '/' {
		segments[0] = segments[0][:0]
		return 1
	}

	updateseg := func(segment []byte, j int) ([]byte, int) {
		segments[j] = segment
		j++
		return segments[j][:0], j
	}

	var ch rune

	u, segment, escape := [6]byte{}, segments[j][:0], false
	for _, ch = range bytes2str(in[1:]) {
		if ch == '~' {
			escape = true

		} else if escape {
			switch ch {
			case '1':
				segment = append(segment, '/')
			case '0':
				segment = append(segment, '~')
			}
			escape = false

		} else if ch == '/' {
			segment, j = updateseg(segment, j)

		} else if ch < utf8.RuneSelf {
			segment = append(segment, byte(ch))

		} else {
			sz := utf8.EncodeRune(u[:], ch)
			segment = append(segment, u[:sz]...)
		}
	}
	if len(segment) > 0 {
		segment, j = updateseg(segment, j)
	}
	if in[len(in)-1] == '/' {
		_, j = updateseg(segment, j)
	}
	return j
}

func encodePointer(p []string, out []byte) int {
	n := 0
	for _, s := range p {
		out[n] = '/'
		n++
		for _, c := range str2bytes(s) {
			switch c {
			case '/':
				out[n] = '~'
				out[n+1] = '1'
				n += 2
			case '~':
				out[n] = '~'
				out[n+1] = '0'
				n += 2
			default:
				out[n] = c
				n++
			}
		}
	}

	// to json string
	var jsonstr [2048]byte
	bs, err := encodeString(out[:n], jsonstr[:0])
	if err != nil {
		panic(err)
	}
	return copy(out, bs[1:len(bs)-1])
}

func allpaths(doc interface{}, pointers []string, prefix []byte) []string {
	var scratch [64]byte

	n := len(prefix)
	prefix = append(prefix, '/', '-')
	switch v := doc.(type) {
	case []interface{}:
		pointers = append(pointers, string(prefix)) // new allocation
		if len(v) > 0 {
			for i, val := range v {
				prefix = prefix[:n]
				dst := strconv.AppendInt(scratch[:0], int64(i), 10)
				prefix = append(prefix, '/')
				prefix = append(prefix, dst...)
				pointers = append(pointers, string(prefix)) // new allocation
				pointers = allpaths(val, pointers, prefix)
			}
		}

	case map[string]interface{}:
		pointers = append(pointers, string(prefix)) // new allocation
		if len(v) > 0 {
			for key, val := range v {
				prefix = prefix[:n]
				prefix = append(prefix, '/')
				prefix = append(prefix, escapeJp(key)...)
				pointers = append(pointers, string(prefix)) // new allocation
				pointers = allpaths(val, pointers, prefix)
			}
		}

	case [][2]interface{}:
		pointers = append(pointers, string(prefix)) // new allocation
		if len(v) > 0 {
			for _, pairs := range v {
				prefix = prefix[:n]
				key, val := pairs[0].(string), pairs[1]
				prefix = append(prefix, '/')
				prefix = append(prefix, escapeJp(key)...)
				pointers = append(pointers, string(prefix)) // new allocation
				pointers = allpaths(val, pointers, prefix)
			}
		}

	}
	return pointers
}

func escapeJp(key string) string {
	if strings.ContainsAny(key, "~/") {
		return strings.Replace(strings.Replace(key, "~", "~0", -1), "/", "~1", -1)
	}
	return key
}
