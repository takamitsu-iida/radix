package radix

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net/netip"
	"testing"
)

var keys = []string{
	"",
	"あ",
	"い",
	"う",
	"え",
	"お",
	"あい",
	"あいう",
	"あいうえ",
	"あいうえお",
	"あかさたな",
}

func TestInsertDelete(t *testing.T) {
	r := New()

	// insert
	for i, key := range keys {
		r.Insert(key, i)
	}

	// check length
	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	// go test ./radix -v
	// print the tree
	r.Walk(func(k string, v interface{}) bool {
		fmt.Println(k, v)
		return false
	})

	// delete
	for _, key := range keys {
		_, ok := r.Delete(key)
		if !ok {
			t.Fatalf("delete failed %q", key)
		}
	}

	// check length
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}
}

func TestUpdate(t *testing.T) {
	// create duplicated keys
	duplicatedKeys := append(keys, keys...)

	r := New()

	// insert
	for i, key := range duplicatedKeys {
		r.Insert(key, i)
	}

	// check length
	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	// delete
	for _, s := range keys {
		r.Delete(s)
	}

	// check length
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}
}

func TestUndelete(t *testing.T) {
	r := New()

	// insert
	for i, key := range keys {
		r.Insert(key, i)
	}

	// delete the keys not in the tree
	for _, key := range keys {
		_, deleted := r.Delete(key + "_dummy")
		if deleted {
			t.Fatalf("delete unexpected, %v", key+"_dummy")
		}
	}

	// check length
	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	// delete
	for _, key := range keys {
		value, deleted := r.Delete(key)
		if !deleted {
			fmt.Println("key=", key, "value=", value, "deleted successfully")
		}
	}

	// check length
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}

}

func TestLongestMatch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a", ""},
		{"ab", ""},
		{"abc", ""},
		{"あ", "あ"},
		{"あいう", "あいう"},
		{"あいかわ", "あい"},
		{"あいみょん", "あい"},
		{"あいうえおかきくけこ", "あいうえお"},
		{"あした", "あ"},
		{"あか", "あ"},
		{"あお", "あ"},
	}

	r := New()

	for _, key := range keys {
		r.Insert(key, nil)
	}

	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	for _, tt := range tests {
		m, _, ok := r.LongestMatch(tt.input)
		if !ok {
			t.Fatalf("key not found: %v", tt.input)
		}
		if m != tt.expected {
			t.Fatalf("key=%v, expected=%v, got=%v", tt.input, tt.expected, m)
		}
	}
}

func testUuid(t *testing.T) {
	var min, max string
	m := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		u := uuid()
		m[u] = i
		if u < min || i == 0 {
			min = u
		}
		if u > max || i == 0 {
			max = u
		}
	}

	r := New()
	r.Load(m)

	if r.Len() != len(m) {
		t.Fatalf("expected length=%v, got=%v", r.Len(), len(m))
	}

	// print the tree
	r.Walk(func(k string, v interface{}) bool {
		fmt.Println(k, v)
		return false
	})

	// check v == Get(k)
	for k, v := range m {
		value, ok := r.Get(k)
		if !ok {
			t.Fatalf("key not found: %v", k)
		}
		if value != v {
			t.Fatalf("expected=%v, got=%v", v, value)
		}
	}

	// check top of the tree
	top, _, _ := r.Top()
	if top != min {
		t.Fatalf("expected top value=%v, got=%v", min, top)
	}

	// check bottom of the tree
	bottom, _, _ := r.Bottom()
	if bottom != max {
		t.Fatalf("expected bottom=%v, got=%v", max, bottom)
	}

	// delete
	for k, v := range m {
		value, ok := r.Delete(k)
		if !ok {
			t.Fatalf("key not found: %v", k)
		}
		if value != v {
			t.Fatalf("expected=%v, got=%v", v, value)
		}
	}

	// check length
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}
}

// generate psuedo uuid
func uuid() string {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func cidrToBinaryString(s string) (string, int, error) {
	prefix, err := netip.ParsePrefix(s)
	if err != nil {
		return "", 0, err
	}

	addrs := prefix.Addr().As4()
	maskLen := prefix.Bits()

	var out bytes.Buffer
	for i := 0; i < 4; i++ {
		out.WriteString(fmt.Sprintf("%08b", addrs[i]))
	}

	return out.String(), maskLen, nil
}

func addrToBinaryString(s string) (string, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return "", err
	}
	addrs := addr.As4()

	var out bytes.Buffer
	for i := 0; i < 4; i++ {
		out.WriteString(fmt.Sprintf("%08b", addrs[i]))
	}

	return out.String(), nil
}

func TestIP(t *testing.T) {
	// routing table
	routes := []struct {
		prefix  string
		gateway string
	}{
		{"10.0.0.0/8", "gig1"},
		{"10.0.0.0/16", "gig2"},
		{"10.0.0.0/24", "gig3"},
		{"192.168.0.0/24", "gig4"},
		{"192.168.0.128/25", "gig5"},
	}

	r := New()

	// convert prefix to a bit string, then insert into radix tree
	for _, route := range routes {
		addr, masklen, err := cidrToBinaryString(route.prefix)
		if err != nil {
			t.Fatalf("failed to convert string: %v", route.prefix)
		}
		addr = addr[:masklen]
		fmt.Println(addr)
		r.Insert(addr, route.gateway)
	}

	tests := []struct {
		destination string
		expected    string
	}{
		{"10.0.0.1", "gig3"},
		{"10.0.1.1", "gig2"},
		{"10.1.1.1", "gig1"},
		{"192.168.0.1", "gig4"},
		{"192.168.0.129", "gig5"},
	}

	for _, test := range tests {
		addr, err := addrToBinaryString(test.destination)
		if err != nil {
			t.Fatal("failed to convert string", err)
		}

		_, v, found := r.LongestMatch(addr)
		if found == false {
			t.Fatalf("key not found: %v", test.destination)
		}
		if test.expected != v {
			t.Fatalf("expected: %v, got: %v", test.expected, v)
		}
	}
}
