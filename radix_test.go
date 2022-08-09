package radix

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
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

func ipToUint(ip net.IP) uint32 {
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func cidrToBinaryString(s string) string {
	ipv4Addr, ipv4Net, err := net.ParseCIDR(s)
	if err != nil {
		return ""
	}
	addr4 := ipv4Addr.To4()
	maskLen, _ := ipv4Net.Mask.Size()

	var out bytes.Buffer
	for i := 0; i < 4; i++ {
		out.WriteString(fmt.Sprintf("%08b", addr4[i]))
	}
	out.WriteString("/")
	out.WriteString(fmt.Sprintf("%d", maskLen))
	return out.String()
}

func TestIP(t *testing.T) {
	s := cidrToBinaryString("10.1.0.0/24")
	fmt.Println(s)
	s = cidrToBinaryString("10.1.0.0/25")
	fmt.Println(s)
}

func TestInsertDelete(t *testing.T) {
	r := New()

	// 挿入
	for i, key := range keys {
		r.Insert(key, i)
	}

	// 長さチェック
	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	// go test ./radix -v
	// ツリーを表示
	r.Walk(func(k string, v interface{}) bool {
		fmt.Println(k, v)
		return false
	})

	// 削除
	for _, key := range keys {
		_, ok := r.Delete(key)
		if !ok {
			t.Fatalf("delete failed %q", key)
		}
	}

	// 長さチェック
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}
}

func TestUpdate(t *testing.T) {
	// keysを連結して重複したキーを作り出す
	duplicatedKeys := append(keys, keys...)

	r := New()

	// 挿入
	for i, key := range duplicatedKeys {
		r.Insert(key, i)
	}

	// 長さチェック
	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	// go test ./radix -v
	// ツリーを表示
	r.Walk(func(k string, v interface{}) bool {
		fmt.Println(k, v)
		return false
	})

	// 削除
	for _, s := range keys {
		r.Delete(s)
	}

	// 長さチェック
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}
}

func TestUndelete(t *testing.T) {
	r := New()

	// keysを挿入
	for i, key := range keys {
		r.Insert(key, i)
	}

	// keysではないものを削除
	for _, key := range keys {
		_, deleted := r.Delete(key + "_dummy")
		if deleted {
			t.Fatalf("delete unexpected, %v", key+"_dummy")
		}
	}

	// 長さチェック
	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

	// keysを削除
	for _, key := range keys {
		value, deleted := r.Delete(key)
		if !deleted {
			fmt.Println("key=", key, "value=", value, "deleted successfully")
		}
	}

	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}

}

func TestLongestMatch(t *testing.T) {
	r := New()

	for _, key := range keys {
		r.Insert(key, nil)
	}

	if r.Len() != len(keys) {
		t.Fatalf("expected length=%v, got=%v", len(keys), r.Len())
	}

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

	// 表示する
	r.Walk(func(k string, v interface{}) bool {
		fmt.Println(k, v)
		return false
	})

	// v == Get(k)を確認
	for k, v := range m {
		value, ok := r.Get(k)
		if !ok {
			t.Fatalf("key not found: %v", k)
		}
		if value != v {
			t.Fatalf("expected=%v, got=%v", v, value)
		}
	}

	// ツリーのトップ
	top, _, _ := r.Top()
	if top != min {
		t.Fatalf("expected top value=%v, got=%v", min, top)
	}

	// ツリーのボトム
	bottom, _, _ := r.Bottom()
	if bottom != max {
		t.Fatalf("expected bottom=%v, got=%v", max, bottom)
	}

	// 一つずつツリーから削除する
	for k, v := range m {
		value, ok := r.Delete(k)
		if !ok {
			t.Fatalf("key not found: %v", k)
		}
		if value != v {
			t.Fatalf("expected=%v, got=%v", v, value)
		}
	}

	// 全部消してゼロになったか
	if r.Len() != 0 {
		t.Fatalf("expected length=%v, got=%v", 0, r.Len())
	}
}

// 乱数を使った疑似UUIDの生成
func uuid() string {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
