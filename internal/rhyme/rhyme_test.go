package rhyme

import (
	"testing"

	"github.com/abkvme/seed-poet/v2/data"
)

func TestBuildPhoneticIndex(t *testing.T) {
	idx := BuildPhoneticIndex()
	if len(idx.Groups) != len(data.Rhymes) {
		t.Fatalf("groups = %d, want %d", len(idx.Groups), len(data.Rhymes))
	}
	keys := idx.GetValidKeys()
	if len(keys) == 0 {
		t.Fatal("no usable rhyme groups")
	}
	for i, k := range keys {
		g := idx.GetGroup(k)
		if len(g) < 2 {
			t.Errorf("key %q has %d words, want >= 2", k, len(g))
		}
		if i > 0 {
			prev := idx.GetGroup(keys[i-1])
			if len(prev) < len(g) || (len(prev) == len(g) && keys[i-1] > k) {
				t.Errorf("keys not ordered by size desc, then name: %q before %q", keys[i-1], k)
			}
		}
	}
	if idx.GetGroup("no-such-key") != nil {
		t.Error("unknown key should return nil")
	}
}

func TestBuildPhoneticIndexDeterministic(t *testing.T) {
	a, b := BuildPhoneticIndex().GetValidKeys(), BuildPhoneticIndex().GetValidKeys()
	if len(a) != len(b) {
		t.Fatal("key count differs between builds")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("key %d differs: %q vs %q", i, a[i], b[i])
		}
	}
}
