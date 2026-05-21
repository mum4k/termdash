// Copyright 2026 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package threed

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png" // register PNG decoder for downloaded emoji assets
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// openMojiFetchURLs lists CDN URLs to try in order when fetching an OpenMoji
// color PNG. Filenames use uppercase hex codepoints joined by hyphens
// (e.g. "1F600.png"), matching bundledEmojiAssetName exactly.
//
// The npm jsdelivr URL is tried first; the GitHub jsdelivr mirror is the
// fallback in case of a package-not-found or stale CDN edge.
var openMojiFetchURLs = []string{
	"https://cdn.jsdelivr.net/npm/openmoji@15.0.0/color/72x72/",
	"https://cdn.jsdelivr.net/gh/hfg-gmuend/openmoji@15.0.0/color/72x72/",
}

// fetchClient is shared across calls; the 5-second timeout prevents the
// animation loop from hanging if the CDN is unreachable.
var fetchClient = &http.Client{Timeout: 5 * time.Second}

// emojiCacheDir returns the OS-appropriate directory used to persist
// downloaded OpenMoji PNGs between process lifetimes.
//
//	macOS:  ~/Library/Caches/termdash-emoji/
//	Linux:  ~/.cache/termdash-emoji/
//
// Returns "" when os.UserCacheDir is unavailable.
func emojiCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "termdash-emoji")
}

// rasterizeSymbolMask resolves an OpenMoji PNG for frame, decodes it, and
// converts the result into a glyphMask.
//
// Resolution order:
//  1. OS disk cache  – avoids CDN traffic and macOS firewall prompts on
//     every run after the first successful download.
//  2. CDN (two mirrors) – downloads and then writes to the disk cache so
//     subsequent calls skip the network entirely.
//
// The in-memory rasterMaskCache (in glyphmask.go) is the outermost layer;
// this function is called at most once per unique emoji string per process.
func rasterizeSymbolMask(frame string, resolution int) (glyphMask, error) {
	name := bundledEmojiAssetName(frame)
	if name == "" {
		return glyphMask{}, fmt.Errorf("rasterizeSymbolMask: cannot derive asset name for %q", frame)
	}

	// 1. Try the disk cache first.
	if mask, ok := loadDiskCachedMask(name, resolution); ok {
		return mask, nil
	}

	// 2. Fetch from CDN, persist to disk, decode.
	var lastErr error
	for _, base := range openMojiFetchURLs {
		url := base + name + ".png"
		mask, err := fetchDecodeAndCache(url, name, resolution)
		if err == nil {
			return mask, nil
		}
		lastErr = err
	}
	return glyphMask{}, lastErr
}

// loadDiskCachedMask reads a previously downloaded emoji PNG from the OS
// cache directory and converts it to a glyphMask at the requested resolution.
func loadDiskCachedMask(name string, resolution int) (glyphMask, bool) {
	dir := emojiCacheDir()
	if dir == "" {
		return glyphMask{}, false
	}
	f, err := os.Open(filepath.Join(dir, name+".png"))
	if err != nil {
		return glyphMask{}, false
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return glyphMask{}, false
	}
	return rasterImageMask(img, resolution), true
}

// fetchDecodeAndCache performs a single HTTP GET, saves the raw PNG bytes to
// the OS cache directory so future process launches can skip the network, and
// returns the decoded glyphMask.
func fetchDecodeAndCache(url, name string, resolution int) (glyphMask, error) {
	resp, err := fetchClient.Get(url)
	if err != nil {
		return glyphMask{}, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return glyphMask{}, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return glyphMask{}, fmt.Errorf("read body from %s: %w", url, err)
	}

	// Persist to disk so the next run reads locally instead of hitting the CDN.
	saveDiskCachedMask(name, data)

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return glyphMask{}, fmt.Errorf("decode PNG from %s: %w", url, err)
	}
	return rasterImageMask(img, resolution), nil
}

// saveDiskCachedMask writes raw PNG bytes to the OS cache directory.
// Errors are silently ignored — the cache is purely a performance optimisation
// and must not break the render path if the filesystem is read-only or full.
func saveDiskCachedMask(name string, data []byte) {
	dir := emojiCacheDir()
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, name+".png"), data, 0o644)
}
