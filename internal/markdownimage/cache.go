package markdownimage

import (
	"container/list"
	"context"
	"crypto/sha256"
	"sync"

	"github.com/egoist/mygo/ui"
)

const maxEntries = 64
const maxDecodedBytes = 64 << 20
const maxPendingBytes = 32 << 20
const workers = 4

type Result struct {
	Bitmap *ui.Bitmap
	Key    [32]byte
	Owner  uint64
	Done   bool
	Err    error
}
type entry struct {
	key    [32]byte
	result Result
	cost   int64
	node   *list.Element
}
type job struct {
	raw    string
	key    [32]byte
	entry  *entry
	epoch  uint64
	ctx    context.Context
	notify func()
}
type loader func(context.Context, string) (*ui.Bitmap, error)

// Cache owns only remote/data image entries. Toolkit file-image caching remains
// unchanged. Keys are hashes so completed entries never retain data URLs/tokens.
type Cache struct {
	start          sync.Once
	mu             sync.Mutex
	entries        map[[32]byte]*entry
	order          list.List
	bytes, pending int64
	epoch          uint64
	ctx            context.Context
	cancel         context.CancelFunc
	lifetime       context.Context
	stop           context.CancelFunc
	jobs           chan job
	load           loader
	closeTransport func()
	closed         bool
}

func New() *Cache {
	client := newClient()
	c := newCache(func(ctx context.Context, raw string) (*ui.Bitmap, error) { return load(ctx, client, raw) })
	c.closeTransport = client.CloseIdleConnections
	return c
}
func newCache(load loader) *Cache {
	life, stop := context.WithCancel(context.Background())
	ctx, cancel := context.WithCancel(life)
	c := &Cache{entries: map[[32]byte]*entry{}, lifetime: life, stop: stop, ctx: ctx, cancel: cancel, jobs: make(chan job, maxEntries), load: load}

	return c
}
func (c *Cache) Request(raw string, notify func()) Result {
	if len(raw) > ((MaxBytes+2)/3)*4+128 {
		return Result{Done: true, Err: ErrLimit}
	}
	raw = Normalize(raw)
	key := sha256.Sum256([]byte(raw))
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return Result{Done: true, Err: context.Canceled}
	}
	c.start.Do(func() {
		for range workers {
			go c.worker()
		}
	})
	if e := c.entries[key]; e != nil {
		c.order.MoveToFront(e.node)
		return e.result
	}
	if int64(len(raw))+c.pending > maxPendingBytes {
		return Result{Done: true, Err: ErrLimit}
	}
	for len(c.entries) >= maxEntries {
		if !c.evict() {
			return Result{Done: true, Err: ErrLimit}
		}
	}
	e := &entry{key: key, result: Result{Key: key, Owner: c.epoch}}
	e.node = c.order.PushFront(e)
	c.entries[key] = e
	j := job{raw: raw, key: key, entry: e, epoch: c.epoch, ctx: c.ctx, notify: notify}
	select {
	case c.jobs <- j:
		c.pending += int64(len(raw))
	default:
		c.remove(e)
		return Result{Done: true, Err: ErrLimit}
	}
	return e.result
}
func (c *Cache) remove(e *entry) { delete(c.entries, e.key); c.order.Remove(e.node); c.bytes -= e.cost }
func (c *Cache) evict() bool {
	for n := c.order.Back(); n != nil; n = n.Prev() {
		e := n.Value.(*entry)
		if e.result.Done {
			c.remove(e)
			return true
		}
	}
	return false
}
func (c *Cache) worker() {
	for {
		select {
		case <-c.lifetime.Done():
			return
		case j := <-c.jobs:
			bitmap, err := c.load(j.ctx, j.raw)
			if err != nil {
				bitmap = nil
			}
			c.mu.Lock()
			if !c.closed && j.epoch == c.epoch {
				c.pending -= int64(len(j.raw))
			}
			live := !c.closed && j.epoch == c.epoch && c.entries[j.key] == j.entry && j.ctx.Err() == nil
			if live {
				var cost int64
				if bitmap != nil {
					w, h := bitmap.Size()
					cost = int64(w) * int64(h) * 6
				} // RGBA plus bounded mip levels
				for c.bytes+cost > maxDecodedBytes {
					if !c.evict() {
						bitmap = nil
						cost = 0
						err = ErrLimit
						break
					}
				}
				j.entry.result = Result{Bitmap: bitmap, Key: j.key, Owner: j.epoch, Done: true, Err: err}
				j.entry.cost = cost
				c.bytes += cost
			}
			c.mu.Unlock()
			if live && j.notify != nil {
				j.notify()
			}
		}
	}
}

// Reset cancels the previous owner even for A→B→A before another render frame.
func (c *Cache) Reset() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.cancel()
	c.epoch++
	c.entries = map[[32]byte]*entry{}
	c.order.Init()
	c.bytes, c.pending = 0, 0
	c.drain()
	c.ctx, c.cancel = context.WithCancel(c.lifetime)
}
func (c *Cache) drain() {
	for {
		select {
		case <-c.jobs:
		default:
			return
		}
	}
}
func (c *Cache) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.cancel()
	c.stop()
	c.drain()
	c.entries = nil
	c.order.Init()
	c.bytes, c.pending = 0, 0
	c.mu.Unlock()
	if c.closeTransport != nil {
		c.closeTransport()
	}
}
