package markdownimage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func raster(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if e := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 32, 16))); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func dataURL(media string, data []byte) string {
	return "data:" + media + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// This dial capability exists only in this test file. Production New has no
// custom-client/private-network override and always uses publicDial.
func controlledClient(server *httptest.Server) *http.Client {
	c := newClient()
	c.Transport = &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}}
	return c
}
func TestAddressValidationAndExactResolvedDial(t *testing.T) {
	blocked := []string{"127.0.0.1", "10.0.0.1", "172.16.1.2", "192.168.1.1", "169.254.169.254", "168.63.129.16", "2001:4860:4860::8888%eth0", "0.1.2.3", "100.64.0.1", "224.0.0.1", "240.0.0.1", "::1", "::", "fe80::1", "fc00::1", "fec0::1", "::ffff:127.0.0.1", "::ffff:10.0.0.1", "64:ff9b::a00:1", "2002:7f00:1::"}
	for _, ip := range blocked {
		if publicIP(netip.MustParseAddr(ip)) {
			t.Errorf("unsafe resolved IP accepted: %s", ip)
		}
	}
	for _, raw := range []string{"file:///etc/passwd", "javascript:alert(1)", "http://user:SECRET@example.com/x", "http://localhost/x", "http://host.local/x", "http://[::1]/x", "http://[::ffff:127.0.0.1]/x"} {
		if _, e := validateURL(raw); e == nil {
			t.Errorf("unsafe URL accepted: %s", raw)
		}
	}
	called := 0
	var dialed string
	lookup := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("::ffff:8.8.8.8")}, nil
	}
	dial := func(_ context.Context, network, address string) (net.Conn, error) {
		called++
		dialed = address
		a, b := net.Pipe()
		b.Close()
		return a, nil
	}
	conn, e := publicDial(lookup, dial)(context.Background(), "tcp", "image.example:443")
	if e != nil {
		t.Fatal(e)
	}
	conn.Close()
	if called != 1 || dialed != "8.8.8.8:443" {
		t.Fatal("not dialing the validated exact address", called, dialed)
	}
	lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")}, nil
	}
	if _, e = publicDial(lookup, dial)(context.Background(), "tcp", "image.example:443"); e == nil || called != 1 {
		t.Fatal("mixed DNS answers bypassed policy", e, called)
	}
}
func TestControlledHTTPRedirectsNeverForwardCredentials(t *testing.T) {
	pngData := raster(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		for _, name := range []string{"Authorization", "Cookie", "Referer"} {
			if r.Header.Get(name) != "" {
				t.Errorf("credential header propagated: %s", name)
			}
		}
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "http://second.example/image", http.StatusFound)
		case "/private":
			http.Redirect(w, r, "http://127.0.0.1/image", http.StatusFound)
		case "/userinfo":
			http.Redirect(w, r, "http://user:SECRET@second.example/image", http.StatusFound)
		case "/scheme":
			http.Redirect(w, r, "file:///etc/passwd", http.StatusFound)
		case "/loop":
			http.Redirect(w, r, "http://image.example/loop", http.StatusFound)
		default:
			w.Header().Set("Content-Type", "image/png")
			w.Write(pngData)
		}
	}))
	defer server.Close()
	client := controlledClient(server)
	defer client.CloseIdleConnections()
	bitmap, e := load(context.Background(), client, "http://image.example/redirect?token=SECRET")
	if e != nil || bitmap == nil {
		t.Fatal(e)
	}
	if requests.Load() != 2 {
		t.Fatal("redirect not actually followed", requests.Load())
	}
	for _, path := range []string{"private", "userinfo", "scheme"} {
		before := requests.Load()
		if _, e = load(context.Background(), client, "http://image.example/"+path); e == nil || requests.Load() != before+1 {
			t.Fatal("unsafe hop was sent", path, e, requests.Load()-before)
		}
	}
	before := requests.Load()
	if _, e = load(context.Background(), client, "http://image.example/loop"); e == nil || requests.Load()-before > 5 {
		t.Fatal("redirects unbounded", e, requests.Load()-before)
	}
	before = requests.Load()
	if _, e = load(context.Background(), client, "http://user:SECRET@image.example/image"); e == nil || requests.Load() != before {
		t.Fatal("userinfo reached transport")
	}
	if strings.Contains(e.Error(), "SECRET") {
		t.Fatal("error exposes credential")
	}
	fresh := newClient()
	defer fresh.CloseIdleConnections()
	if fresh.Jar != nil || fresh.Transport.(*http.Transport).Proxy != nil {
		t.Fatal("ambient cookies/proxy enabled")
	}
	req, _ := http.NewRequest("GET", "http://second.example/image", nil)
	req.Header.Set("Cookie", "secret")
	req.Header.Set("Authorization", "secret")
	req.Header.Set("Referer", "secret")
	if e := checkRedirect(req, []*http.Request{{URL: &url.URL{Scheme: "http"}}}); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"Cookie", "Authorization", "Referer"} {
		if req.Header.Get(name) != "" {
			t.Fatal("unsafe hop header", name)
		}
	}
	if e := checkRedirect(req, []*http.Request{{URL: &url.URL{Scheme: "https"}}}); e == nil {
		t.Fatal("HTTPS downgrade accepted")
	}
}
func TestBodyMIMEPixelsAndAnimationAreBounded(t *testing.T) {
	data := raster(t)
	client := newClient()
	defer client.CloseIdleConnections()
	if b, e := load(context.Background(), client, dataURL("image/png", data)); e != nil || b == nil {
		t.Fatal(e)
	}
	for _, raw := range []string{dataURL("image/svg+xml", []byte(`<svg onload="run()"><script>bad()</script></svg>`)), dataURL("image/png", []byte(`<svg/>`)), dataURL("image/jpeg", data), "data:image/png,not-base64"} {
		if b, e := load(context.Background(), client, raw); e == nil || b != nil {
			t.Fatal("unsafe format accepted")
		}
	}
	huge := append([]byte(nil), data...)
	binary.BigEndian.PutUint32(huge[16:20], MaxDimension+1)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	if _, e := decode(huge, "image/png"); !errors.Is(e, ErrLimit) {
		t.Fatal("dimensions checked too late", e)
	}
	if dimensions(int(^uint(0)>>1), int(^uint(0)>>1)) {
		t.Fatal("overflow dimensions accepted")
	}
	if _, e := load(context.Background(), client, "data:image/png;base64,"+strings.Repeat("A", base64.StdEncoding.EncodedLen(MaxBytes)+4)); !errors.Is(e, ErrLimit) {
		t.Fatal("data URI length unbounded", e)
	}
	pal := color.Palette{color.Black, color.White}
	first := image.NewPaletted(image.Rect(0, 0, 2, 1), pal)
	second := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	var encoded bytes.Buffer
	if e := gif.EncodeAll(&encoded, &gif.GIF{Image: []*image.Paletted{first, second}, Delay: []int{1, 1}, Config: image.Config{ColorModel: pal, Width: 4, Height: 4}}); e != nil {
		t.Fatal(e)
	}
	b, e := decode(encoded.Bytes(), "image/gif")
	if e != nil {
		t.Fatal(e)
	}
	if w, h := b.Size(); w != 2 || h != 1 {
		t.Fatal("GIF did not remain first-frame-only", w, h)
	}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "image/png")
		if r.URL.Path == "/declared" {
			w.Header().Set("Content-Length", strconv.Itoa(MaxBytes+1))
			w.WriteHeader(200)
			return
		}
		w.(http.Flusher).Flush()
		chunk := make([]byte, 8192)
		for n := 0; n < MaxBytes/len(chunk)+1; n++ {
			if _, e := w.Write(chunk); e != nil {
				return
			}
		}
	}))
	defer server.Close()
	c := controlledClient(server)
	defer c.CloseIdleConnections()
	for _, path := range []string{"declared", "stream"} {
		if _, e := load(context.Background(), c, "http://image.example/"+path); !errors.Is(e, ErrLimit) {
			t.Fatal("HTTP body not bounded", path, e)
		}
	}
}
func TestControlledHTTPContextCancellation(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.(http.Flusher).Flush()
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	client := controlledClient(server)
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, e := load(ctx, client, "http://image.example/image"); done <- e }()
	<-entered
	cancel()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatal("cancelled request did not preserve cancellation", e)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel did not terminate body read")
	}
}
func TestNormalizeSignedGitHubURLWithoutSendingRequest(t *testing.T) {
	raw := "https://private-user-images.githubusercontent.com/1/2-12345678-1234-1234-1234-123456789abc.png?jwt=SECRET"
	want := "https://github.com/user-attachments/assets/12345678-1234-1234-1234-123456789abc"
	if Normalize(raw) != want {
		t.Fatal("source signed-image normalization differs")
	}
	for _, raw := range []string{"https://private-user-images.githubusercontent.com/1/not-a-uuid.png?jwt=SECRET", "https://private-user-images.githubusercontent.com.evil/1/2-12345678-1234-1234-1234-123456789abc.png"} {
		if Normalize(raw) != raw {
			t.Fatal("rewrote malformed/lookalike origin")
		}
	}
}

func TestAnimatedWebPIsRejectedIncludingAfterUnknownChunk(t *testing.T) {
	for _, prefix := range [][]byte{nil, []byte("JUNK\x00\x00\x00\x00")} {
		data := append([]byte("RIFF\x00\x00\x00\x00WEBP"), prefix...)
		data = append(data, []byte("VP8X\x0a\x00\x00\x00\x02\x00\x00\x00\x01\x00\x00\x01\x00\x00")...)
		binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))
		if !animatedWebP(data) {
			t.Fatal("animation not recognized")
		}
		if b, e := decode(data, "image/webp"); b != nil || !errors.Is(e, ErrUnsafe) {
			t.Fatal("animated WebP accepted", e)
		}
	}
}
