// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cespare/xxhash"
)

var storageTestLock sync.RWMutex

func bkey(i int) string {
	return fmt.Sprintf("%09d", i)
}

func bval(i int) []byte {
	return []byte(fmt.Sprintf("%025d", i))
}

func Test_Put(t *testing.T) {
	s := New(0)

	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}
}

func Test_Get(t *testing.T) {
	s := New(0)

	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}

	for i := 0; i < 100; i++ {
		hkey := xxhash.Sum64([]byte(bkey(i)))
		vdata, err := s.Get(hkey)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
		if vdata.Key != bkey(i) {
			t.Fatalf("Expected %s. Got %s", bkey(i), vdata.Key)
		}
		if vdata.TTL != int64(i) {
			t.Fatalf("Expected %d. Got %v", i, vdata.TTL)
		}
		if !bytes.Equal(vdata.Value, bval(i)) {
			t.Fatalf("Value is malformed for %d", i)
		}
	}
}

func Test_Delete(t *testing.T) {
	s := New(0)

	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}

	for i := 0; i < 100; i++ {
		hkey := xxhash.Sum64([]byte(bkey(i)))
		s.Delete(hkey)
		_, err := s.Get(hkey)
		if err != ErrKeyNotFound {
			t.Fatalf("Expected ErrKeyNotFound. Got: %v", err)
		}
	}

	for _, item := range s.tables {
		if item.inuse != 0 {
			t.Fatal("inuse is different than 0.")
		}
		if len(item.hkeys) != 0 {
			t.Fatalf("Expected key count is zero. Got: %d", len(item.hkeys))
		}
	}
}

func Test_CompactTables(t *testing.T) {
	s := New(0)

	compaction := func() {
		storageTestLock.Lock()
		defer storageTestLock.Unlock()
		for {
			if done := s.CompactTables(); done {
				return
			}
		}
	}
	// Current free space is 1MB. Trigger a compaction operation.
	for i := 0; i < 1500; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: []byte(fmt.Sprintf("%01000d", i)),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))

		storageTestLock.Lock()
		err := s.Put(hkey, vdata)
		storageTestLock.Unlock()

		if err == ErrFragmented {
			go compaction()
			err = nil
		}
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}

	for i := 0; i < 1500; i++ {
		hkey := xxhash.Sum64([]byte(bkey(i)))

		storageTestLock.RLock()
		vdata, err := s.Get(hkey)
		storageTestLock.RUnlock()

		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
		if vdata.Key != bkey(i) {
			t.Fatalf("Expected %s. Got %s", bkey(i), vdata.Key)
		}
		if vdata.TTL != int64(i) {
			t.Fatalf("Expected %d. Got %v", i, vdata.TTL)
		}
		val := []byte(fmt.Sprintf("%01000d", i))
		if !bytes.Equal(vdata.Value, val) {
			t.Fatalf("Value is malformed for %d", i)
		}
	}

	for i := 0; i < 10; i++ {
		storageTestLock.Lock()
		if len(s.tables) == 1 {
			storageTestLock.Unlock()
			// It's OK.
			return
		}
		storageTestLock.Unlock()
		<-time.After(100 * time.Millisecond)
	}
	t.Error("Tables cannot be compacted.")
}

func Test_PurgeTables(t *testing.T) {
	s := New(0)

	compaction := func() {
		storageTestLock.Lock()
		defer storageTestLock.Unlock()
		for {
			if done := s.CompactTables(); done {
				return
			}
		}
	}
	// Current free space is 1MB. Trigger a compaction operation.
	for i := 0; i < 2000; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: []byte(fmt.Sprintf("%01000d", i)),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))

		// Simulate the real-world case.
		storageTestLock.Lock()
		err := s.Put(hkey, vdata)
		storageTestLock.Unlock()

		if err == ErrFragmented {
			go compaction()
			err = nil
		}
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}

	for i := 0; i < 2000; i++ {
		hkey := xxhash.Sum64([]byte(bkey(i)))

		// Simulate the real-world case.
		storageTestLock.Lock()
		err := s.Delete(hkey)
		storageTestLock.Unlock()
		if err == ErrFragmented {
			go compaction()
			err = nil
		}
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}

	for i := 0; i < 10; i++ {
		storageTestLock.Lock()
		if len(s.tables) == 1 {
			// Only has 1 table with minimum size.
			if s.tables[0].allocated == minimumSize {
				storageTestLock.Unlock()
				return
			}
		}
		storageTestLock.Unlock()
		<-time.After(100 * time.Millisecond)
	}
	t.Fatal("Tables cannot be purged.")
}

func Test_ExportImport(t *testing.T) {
	s := New(0)
	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}
	data, err := s.Export()
	if err != nil {
		t.Fatalf("Expected nil. Got %v", err)
	}
	fresh, err := Import(data)
	if err != nil {
		t.Fatalf("Expected nil. Got %v", err)
	}
	for i := 0; i < 100; i++ {
		hkey := xxhash.Sum64([]byte(bkey(i)))
		vdata, err := fresh.Get(hkey)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
		if vdata.Key != bkey(i) {
			t.Fatalf("Expected %s. Got %s", bkey(i), vdata.Key)
		}
		if vdata.TTL != int64(i) {
			t.Fatalf("Expected %d. Got %v", i, vdata.TTL)
		}
		if !bytes.Equal(vdata.Value, bval(i)) {
			t.Fatalf("Value is malformed for %d", i)
		}
	}
}

func Test_Len(t *testing.T) {
	s := New(0)
	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
	}

	if s.Len() != 100 {
		t.Fatalf("Expected length: 100. Got: %d", s.Len())
	}
}

func Test_Range(t *testing.T) {
	s := New(0)
	hkeys := make(map[uint64]struct{})
	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
		hkeys[hkey] = struct{}{}
	}

	s.Range(func(hkey uint64, vdata *VData) bool {
		if _, ok := hkeys[hkey]; !ok {
			t.Fatalf("Invalid hkey: %d", hkey)
		}
		return true
	})
}

func Test_Check(t *testing.T) {
	s := New(0)
	hkeys := make(map[uint64]struct{})
	for i := 0; i < 100; i++ {
		vdata := &VData{
			Key:   bkey(i),
			TTL:   int64(i),
			Value: bval(i),
		}
		hkey := xxhash.Sum64([]byte(vdata.Key))
		err := s.Put(hkey, vdata)
		if err != nil {
			t.Fatalf("Expected nil. Got %v", err)
		}
		hkeys[hkey] = struct{}{}
	}

	for hkey := range hkeys {
		if !s.Check(hkey) {
			t.Fatalf("hkey could not be found: %d", hkey)
		}
	}
}