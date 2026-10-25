module github.com/charlzyx/orbit

go 1.27.3

require (
	github.com/HugoSmits86/nativewebp v1.3.0
	github.com/egoist/mygo v0.6.4-0.20261025001514-9c69cf829ece
	github.com/egoist/mygo/orbit-terminal v0.0.0
	github.com/fsnotify/fsnotify v1.10.1
	golang.org/x/image v0.46.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/ebitengine/purego v0.11.1 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	golang.org/x/net v0.59.0
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/egoist/mygo/orbit-terminal => ./third_party/mygo-terminal
