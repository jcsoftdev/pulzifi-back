package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/minio"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type SnapshotService struct {
	minioClient *minio.Client
}

func NewSnapshotService(minioClient *minio.Client) *SnapshotService {
	return &SnapshotService{minioClient: minioClient}
}

func (s *SnapshotService) CaptureAndUpload(ctx context.Context, req entities.SnapshotRequest) (*entities.SnapshotResult, error) {
	cleanURL := sanitizeURL(req.URL)
	logger.Info("Starting snapshot capture", zap.String("url", cleanURL))

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-site-isolation-trials", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("window-size", "1366,768"),
	)
	if p := os.Getenv("CHROME_PATH"); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	} else {
		switch runtime.GOOS {
		case "darwin":
			opts = append(opts, chromedp.ExecPath("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"))
		default:
			opts = append(opts, chromedp.ExecPath("/usr/bin/chromium"))
		}
	}
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var inflight int32
	var lastActive time.Time
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev.(type) {
		case *network.EventRequestWillBeSent:
			atomic.AddInt32(&inflight, 1)
			lastActive = time.Now()
		case *network.EventLoadingFinished, *network.EventLoadingFailed:
			atomic.AddInt32(&inflight, -1)
			lastActive = time.Now()
		}
	})
	lastActive = time.Now()

	var buf []byte
	var html string
	var text string
	if err := chromedp.Run(taskCtx,
		network.Enable(),
		security.SetIgnoreCertificateErrors(true),
		emulation.SetUserAgentOverride("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"),
		network.SetExtraHTTPHeaders(network.Headers{
			"Accept-Language": "es-ES,es;q=0.9,en-US;q=0.8,en;q=0.7",
		}),
		chromedp.Navigate(cleanURL),
		chromedp.EmulateViewport(1366, 768),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(c context.Context) error {
			deadline := time.Now().Add(20 * time.Second)
			for time.Now().Before(deadline) {
				var rs string
				if err := chromedp.Evaluate(`document.readyState`, &rs).Do(c); err == nil && rs == "complete" {
					return nil
				}
				time.Sleep(300 * time.Millisecond)
			}
			return fmt.Errorf("timeout waiting for document ready")
		}),
		chromedp.ActionFunc(func(c context.Context) error {
			idle := 1500 * time.Millisecond
			timeout := 25 * time.Second
			deadline := time.Now().Add(timeout)
			for time.Now().Before(deadline) {
				if atomic.LoadInt32(&inflight) == 0 && time.Since(lastActive) >= idle {
					return nil
				}
				time.Sleep(250 * time.Millisecond)
			}
			return fmt.Errorf("timeout waiting for network idle")
		}),
		chromedp.ActionFunc(func(c context.Context) error {
			var prevH int64
			var prevW int64
			stableCount := 0
			timeout := 25 * time.Second
			deadline := time.Now().Add(timeout)
			for time.Now().Before(deadline) {
				var h int64
				var w int64
				var fontsLoaded bool
				var imagesLoaded bool
				var ready string
				_ = chromedp.Evaluate(`document.documentElement ? document.documentElement.scrollHeight : 0`, &h).Do(c)
				_ = chromedp.Evaluate(`document.documentElement ? document.documentElement.scrollWidth : 0`, &w).Do(c)
				_ = chromedp.Evaluate(`document.fonts && document.fonts.status === "loaded"`, &fontsLoaded).Do(c)
				_ = chromedp.Evaluate(`Array.from(document.images).every(img => img.complete)`, &imagesLoaded).Do(c)
				_ = chromedp.Evaluate(`document.readyState`, &ready).Do(c)
				if h == prevH && w == prevW {
					stableCount++
				} else {
					stableCount = 0
					prevH = h
					prevW = w
				}
				if ready == "complete" && fontsLoaded && imagesLoaded && stableCount >= 4 {
					return nil
				}
				time.Sleep(250 * time.Millisecond)
			}
			return fmt.Errorf("timeout waiting for stable layout")
		}),
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.Evaluate(`document.documentElement && document.documentElement.innerText ? document.documentElement.innerText : ""`, &text),
		chromedp.FullScreenshot(&buf, 90),
	); err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	ts := time.Now().Unix()
	imgName := fmt.Sprintf("%s/%d.png", req.PageID, ts)
	htmlName := fmt.Sprintf("%s/%d.html", req.PageID, ts)
	textName := fmt.Sprintf("%s/%d.txt", req.PageID, ts)
	imgReader := bytes.NewReader(buf)
	htmlReader := bytes.NewReader([]byte(html))
	textReader := bytes.NewReader([]byte(text))

	imgURL, err := s.minioClient.Upload(ctx, imgName, imgReader, int64(len(buf)), "image/png")
	if err != nil {
		return nil, fmt.Errorf("failed to upload screenshot: %w", err)
	}
	htmlURL, err := s.minioClient.Upload(ctx, htmlName, htmlReader, int64(len(html)), "text/html; charset=utf-8")
	if err != nil {
		return nil, fmt.Errorf("failed to upload html: %w", err)
	}
	textURL, err := s.minioClient.Upload(ctx, textName, textReader, int64(len(text)), "text/plain; charset=utf-8")
	if err != nil {
		return nil, fmt.Errorf("failed to upload text: %w", err)
	}

	imgHash := sha256.Sum256(buf)
	htmlHash := sha256.Sum256([]byte(html))
	textHash := sha256.Sum256([]byte(text))
	combined := sha256.Sum256(append(append(buf, []byte(html)...), []byte(text)...))
	res := &entities.SnapshotResult{
		PageID:      req.PageID,
		URL:         cleanURL,
		SchemaName:  req.SchemaName,
		ImageURL:    imgURL,
		HTMLURL:     htmlURL,
		TextURL:     textURL,
		ImageHash:   hex.EncodeToString(imgHash[:]),
		HTMLHash:    hex.EncodeToString(htmlHash[:]),
		TextHash:    hex.EncodeToString(textHash[:]),
		ContentHash: hex.EncodeToString(combined[:]),
		Status:      "success",
		CreatedAt:   time.Now(),
	}
	logger.Info(
		"Snapshot captured and uploaded",
		zap.String("page_id", res.PageID),
		zap.String("url", res.URL),
		zap.String("schema_name", res.SchemaName),
		zap.String("image_url", res.ImageURL),
		zap.String("html_url", res.HTMLURL),
		zap.String("text_url", res.TextURL),
		zap.String("image_hash", res.ImageHash),
		zap.String("html_hash", res.HTMLHash),
		zap.String("text_hash", res.TextHash),
		zap.String("content_hash", res.ContentHash),
		zap.String("status", res.Status),
		zap.Time("created_at", res.CreatedAt),
	)
	return res, nil
}

func sanitizeURL(raw string) string {
	u := strings.TrimSpace(raw)
	u = strings.Trim(u, "`\"' ")
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "https://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	if parsed.Host == "" && parsed.Path != "" {
		// Handle cases like "example.com" being in Path
		parsed.Host = parsed.Path
		parsed.Path = ""
	}
	return parsed.String()
}
