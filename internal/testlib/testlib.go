// Package testlib provides common test utilities.
package testlib

import (
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// MustTempDir creates a temporary directory, or dies trying.
func MustTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "testlib_")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("test directory: %q", dir)
	return dir
}

// RemoveIfOk removes the given directory, but only if we have not failed. We
// want to keep the failed directories for debugging.
func RemoveIfOk(t *testing.T, dir string) {
	// Safeguard, to make sure we only remove test directories.
	// This should help prevent accidental deletions.
	if !strings.Contains(dir, "testlib_") {
		panic("invalid/dangerous directory")
	}

	if !t.Failed() {
		os.RemoveAll(dir)
	}
}

// Rewrite a file with the given contents.
func Rewrite(t *testing.T, path, contents string) error {
	// Safeguard, to make sure we only mess with test files.
	if !strings.Contains(path, "testlib_") {
		panic("invalid/dangerous path")
	}

	err := ioutil.WriteFile(path, []byte(contents), 0600)
	if err != nil {
		t.Errorf("failed to rewrite file: %v", err)
	}

	return err
}

// GetFreePort returns a free TCP port. This is hacky and not race-free, but
// it works well enough for testing purposes.
func GetFreePort() string {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func WaitFor(f func() bool, d time.Duration) bool {
	start := time.Now()
	for time.Since(start) < d {
		if f() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

type DeliverRequest struct {
	From string
	To   string
	Data []byte
}

// Courier for test purposes. Never fails, and always remembers everything.
type TestCourier struct {
	wg       sync.WaitGroup
	Requests []*DeliverRequest
	ReqFor   map[string]*DeliverRequest
	sync.Mutex
}

func (tc *TestCourier) Deliver(from string, to string, data []byte) (error, bool) {
	defer tc.wg.Done()
	dr := &DeliverRequest{from, to, data}
	tc.Lock()
	tc.Requests = append(tc.Requests, dr)
	tc.ReqFor[to] = dr
	tc.Unlock()
	return nil, false
}

func (tc *TestCourier) Expect(i int) {
	tc.wg.Add(i)
}

func (tc *TestCourier) Wait() {
	tc.wg.Wait()
}

// NewTestCourier returns a new, empty TestCourier instance.
func NewTestCourier() *TestCourier {
	return &TestCourier{
		ReqFor: map[string]*DeliverRequest{},
	}
}

type dumbCourier struct{}

func (c dumbCourier) Deliver(from string, to string, data []byte) (error, bool) {
	return nil, false
}

// Dumb courier, for when we just don't care about the result.
var DumbCourier = dumbCourier{}
