// Package markdownimage loads untrusted Markdown raster images with bounded
// resources. It never uses ambient cookies, credentials or proxy settings.
package markdownimage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"image"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

const MaxBytes = 8 << 20
const MaxPixels = 8_000_000
const MaxDimension = 8192

var ErrUnavailable = errors.New("image unavailable")
var ErrUnsafe = errors.New("image address or format is not allowed")
var ErrLimit = errors.New("image exceeds resource limit")

// Normalize mirrors Rust transcript_view.rs:5699's durable GitHub asset URL.
// It does not turn other hosts or malformed signed URLs into trusted addresses.
func Normalize(raw string) string {
	const prefix = "https://private-user-images.githubusercontent.com/"
	if !strings.HasPrefix(raw, prefix) {
		return raw
	}
	path := strings.SplitN(strings.SplitN(strings.TrimPrefix(raw, prefix), "?", 2)[0], "#", 2)[0]
	name := path[strings.LastIndex(path, "/")+1:]
	_, tail, ok := strings.Cut(name, "-")
	if !ok {
		return raw
	}
	if i := strings.LastIndex(tail, "."); i >= 0 {
		tail = tail[:i]
	}
	if len(tail) != 36 {
		return raw
	}
	for i, c := range tail {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return raw
			}
		} else if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return raw
		}
	}
	return "https://github.com/user-attachments/assets/" + tail
}
func validateURL(raw string) (*url.URL, error) {
	if len(raw) > 8192 {
		return nil, ErrLimit
	}
	u, e := url.Parse(raw)
	if e != nil || u.User != nil || u.Opaque != "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, ErrUnsafe
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "" || strings.Contains(host, "%") || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return nil, ErrUnsafe
	}
	if ip, e := netip.ParseAddr(host); e == nil && !publicIP(ip) {
		return nil, ErrUnsafe
	}
	u.Fragment = ""
	return u, nil
}

var specialNetworks = []netip.Prefix{
	netip.MustParsePrefix("168.63.129.16/32"), netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("::/96"), netip.MustParsePrefix("64:ff9b::/96"), netip.MustParsePrefix("64:ff9b:1::/48"), netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("fec0::/10"), netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2001::/32"),
}

func publicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.Zone() != "" || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return false
	}
	for _, p := range specialNetworks {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}

type lookupIPs func(context.Context, string, string) ([]netip.Addr, error)
type dialIP func(context.Context, string, string) (net.Conn, error)

// Resolve once, validate every answer and dial that exact vetted IP. The
// hostname is never handed to a second resolver at the connection boundary.
func publicDial(lookup lookupIPs, dial dialIP) dialIP {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, e := net.SplitHostPort(address)
		if e != nil {
			return nil, ErrUnsafe
		}
		ips, e := lookup(ctx, "ip", host)
		if e != nil || len(ips) == 0 {
			return nil, ErrUnavailable
		}
		for _, ip := range ips {
			if !publicIP(ip) {
				return nil, ErrUnsafe
			}
		}
		for _, ip := range ips {
			conn, e := dial(ctx, network, net.JoinHostPort(ip.Unmap().String(), port))
			if e == nil {
				return conn, nil
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}
		return nil, ErrUnavailable
	}
}
func newClient() *http.Client {
	d := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	tr := &http.Transport{Proxy: nil, DialContext: publicDial(net.DefaultResolver.LookupNetIP, d.DialContext), ForceAttemptHTTP2: true,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 5 * time.Second,
		MaxResponseHeaderBytes: 64 << 10, MaxIdleConns: 4, MaxIdleConnsPerHost: 4, MaxConnsPerHost: 4, IdleConnTimeout: 30 * time.Second}
	return &http.Client{Transport: tr, Timeout: 12 * time.Second, CheckRedirect: checkRedirect}
}
func imageHeaders(r *http.Request) {
	r.Header = make(http.Header)
	r.Header.Set("Accept", "image/png,image/jpeg,image/gif,image/webp,image/bmp")
	r.Header.Set("User-Agent", "Orbit-Markdown-Image")
}
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return ErrLimit
	}
	if _, e := validateURL(req.URL.String()); e != nil {
		return e
	}
	if len(via) > 0 && via[len(via)-1].URL.Scheme == "https" && req.URL.Scheme != "https" {
		return ErrUnsafe
	}
	imageHeaders(req) // no Referer, Cookie or Authorization from any prior hop
	return nil
}
func load(ctx context.Context, client *http.Client, raw string) (*ui.Bitmap, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	var data []byte
	var media string
	var e error
	if strings.HasPrefix(strings.ToLower(raw), "data:") {
		head, payload, ok := strings.Cut(raw, ",")
		if !ok || len(head) > 128 || !strings.HasSuffix(strings.ToLower(head), ";base64") {
			return nil, ErrUnsafe
		}
		if len(payload) > base64.StdEncoding.EncodedLen(MaxBytes) || strings.ContainsAny(payload, "\r\n\t ") {
			return nil, ErrLimit
		}
		media = strings.ToLower(head[5 : len(head)-7])
		if !rasterMIME(media) {
			return nil, ErrUnsafe
		}
		data, e = base64.StdEncoding.Strict().DecodeString(payload)
		if e != nil {
			return nil, ErrUnavailable
		}
	} else {
		u, err := validateURL(raw)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, ErrUnsafe
		}
		imageHeaders(req)
		response, err := client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, ErrUnavailable
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, ErrUnavailable
		}
		if response.ContentLength > MaxBytes {
			return nil, ErrLimit
		}
		media, _, e = mime.ParseMediaType(response.Header.Get("Content-Type"))
		if e != nil {
			return nil, ErrUnsafe
		}
		data, e = io.ReadAll(io.LimitReader(response.Body, MaxBytes+1))
		if e != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, ErrUnavailable
		}
	}
	if len(data) > MaxBytes {
		return nil, ErrLimit
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	bitmap, err := decode(data, media)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return bitmap, err
}
func rasterMIME(media string) bool {
	switch media {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp", "image/x-ms-bmp":
		return true
	}
	return false
}
func dimensions(w, h int) bool {
	return w > 0 && h > 0 && w <= MaxDimension && h <= MaxDimension && int64(w) <= int64(MaxPixels)/int64(h)
}
func decode(data []byte, media string) (*ui.Bitmap, error) {
	if !rasterMIME(media) && media != "application/octet-stream" {
		return nil, ErrUnsafe
	}
	// Reject animation containers before asking the decoder for configuration.
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" && animatedWebP(data) {
		return nil, ErrUnsafe
	}
	// SVG is deliberately not parsed, even if advertised as a raster image.
	config, format, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return nil, ErrUnavailable
	}
	if !dimensions(config.Width, config.Height) {
		return nil, ErrLimit
	}
	expected := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "gif": "image/gif", "webp": "image/webp", "bmp": "image/bmp"}[format]
	if expected == "" || (media != "application/octet-stream" && media != expected && !(format == "bmp" && media == "image/x-ms-bmp")) {
		return nil, ErrUnsafe
	}
	// GIF uses only image.Decode's first frame, never DecodeAll.
	bitmap, e := ui.DecodeBitmap(data)
	if e != nil {
		return nil, ErrUnavailable
	}
	w, h := bitmap.Size()
	if !dimensions(w, h) {
		return nil, ErrLimit
	}
	return bitmap, nil
}

// Inspect every bounded RIFF chunk, not only a VP8X at offset 12. Unknown
// chunks are legal and must not conceal an animation flag or frame chunk.
func animatedWebP(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	for offset := 12; offset <= len(data)-8; {
		size := uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		end := uint64(offset) + 8 + size
		if end > uint64(len(data)) {
			return true
		}
		tag := string(data[offset : offset+4])
		if tag == "ANIM" || tag == "ANMF" || (tag == "VP8X" && size > 0 && data[offset+8]&2 != 0) {
			return true
		}
		next := end + (size & 1)
		if next > uint64(len(data)) {
			return true
		}
		offset = int(next)
	}
	return false
}
