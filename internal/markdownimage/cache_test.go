package markdownimage

import (
	"context"
	"errors"
	"fmt"
	"github.com/egoist/mygo/ui"
	"image"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func awaitResult(t *testing.T, c *Cache, raw string) Result {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		r := c.Request(raw, nil)
		if r.Done {
			return r
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("image did not settle")
	return Result{}
}
func TestCacheFailureEntriesAndDecodedBudgetAreBounded(t *testing.T) {
	var calls atomic.Int32
	c := newCache(func(context.Context, string) (*ui.Bitmap, error) { calls.Add(1); return nil, ErrUnavailable })
	defer c.Close()
	for i := 0; i < maxEntries*2; i++ {
		awaitResult(t, c, fmt.Sprint(i))
	}
	before := calls.Load()
	awaitResult(t, c, fmt.Sprint(maxEntries*2-1))
	if calls.Load() != before {
		t.Fatal("failure repeated request")
	}
	c.mu.Lock()
	if len(c.entries) > maxEntries || c.order.Len() != len(c.entries) || c.bytes != 0 {
		t.Error("failure cache unbounded")
	}
	c.mu.Unlock()
	bitmap := ui.NewBitmap(image.NewNRGBA(image.Rect(0, 0, 1024, 1024)))
	pictures := newCache(func(context.Context, string) (*ui.Bitmap, error) { return bitmap, nil })
	defer pictures.Close()
	for i := 0; i < 20; i++ {
		if r := awaitResult(t, pictures, fmt.Sprint(i)); r.Bitmap == nil {
			t.Fatal(r.Err)
		}
	}
	pictures.mu.Lock()
	defer pictures.mu.Unlock()
	if pictures.bytes > maxDecodedBytes || len(pictures.entries) >= 20 {
		t.Fatal("decoded budget not evicted", pictures.bytes, len(pictures.entries))
	}
}
func TestCacheResetDiscardsLateOldOwnerAndCloseCancels(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var calls, oldNotify atomic.Int32
	old := ui.NewBitmap(image.NewNRGBA(image.Rect(0, 0, 1, 1)))
	fresh := ui.NewBitmap(image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	c := newCache(func(ctx context.Context, raw string) (*ui.Bitmap, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-release
			return old, nil
		}
		return fresh, nil
	})
	defer c.Close()
	first := c.Request("same", func() { oldNotify.Add(1) })
	<-started
	c.Reset()
	second := awaitResult(t, c, "same")
	if second.Owner == first.Owner || second.Bitmap != fresh {
		t.Fatal("owner did not advance")
	}
	close(release)
	time.Sleep(10 * time.Millisecond)
	if r := c.Request("same", nil); r.Bitmap != fresh || oldNotify.Load() != 0 {
		t.Fatal("late old owner published")
	}
	c.Close()
	if r := c.Request("same", nil); !errors.Is(r.Err, context.Canceled) {
		t.Fatal("closed cache accepted request")
	}
}
func TestPendingConcurrencyAndCancellationStayBounded(t *testing.T) {
	var active, maxActive, finished atomic.Int32
	entered := make(chan struct{}, workers)
	c := newCache(func(ctx context.Context, raw string) (*ui.Bitmap, error) {
		n := active.Add(1)
		for {
			old := maxActive.Load()
			if n <= old || maxActive.CompareAndSwap(old, n) {
				break
			}
		}
		entered <- struct{}{}
		<-ctx.Done()
		active.Add(-1)
		finished.Add(1)
		return nil, ctx.Err()
	})
	for i := 0; i < maxEntries*2; i++ {
		c.Request(fmt.Sprint(i), nil)
	}
	for range workers {
		<-entered
	}
	c.mu.Lock()
	if len(c.entries) > maxEntries || c.pending > maxPendingBytes {
		t.Error("pending cache unbounded")
	}
	c.mu.Unlock()
	c.Close()
	deadline := time.Now().Add(time.Second)
	for finished.Load() < workers && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if maxActive.Load() > workers || finished.Load() != workers {
		t.Fatal("unbounded or uncancelled workers", maxActive.Load(), finished.Load())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pending != 0 || len(c.entries) != 0 {
		t.Fatal("close retained cache state", c.pending, len(c.entries))
	}
}

func TestPendingEncodedBytesAreBoundedAndReleased(t *testing.T) {
	c := newCache(func(ctx context.Context, raw string) (*ui.Bitmap, error) { <-ctx.Done(); return nil, ctx.Err() })
	defer c.Close()
	accepted, rejected := 0, 0
	for i := 0; i < 5; i++ {
		raw := fmt.Sprint(i) + strings.Repeat("A", 8<<20)
		r := c.Request(raw, nil)
		if errors.Is(r.Err, ErrLimit) {
			rejected++
		} else {
			accepted++
		}
	}
	c.mu.Lock()
	if c.pending > maxPendingBytes || accepted > 3 || rejected < 2 {
		t.Error("encoded request retention unbounded", c.pending, accepted, rejected)
	}
	c.mu.Unlock()
	c.Close()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pending != 0 {
		t.Fatal("close retained pending bytes", c.pending)
	}
}

func TestFailedLoadCannotPublishPartialBitmap(t *testing.T) {
	bitmap := ui.NewBitmap(image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	c := newCache(func(context.Context, string) (*ui.Bitmap, error) { return bitmap, ErrUnsafe })
	defer c.Close()
	r := awaitResult(t, c, "failed")
	if r.Bitmap != nil || !errors.Is(r.Err, ErrUnsafe) {
		t.Fatal("failed loader exposed partial image", r)
	}
}
