package dashboard

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/grigata/chta-network-stats/internal/models"
	"github.com/grigata/chta-network-stats/internal/statistics"
	"github.com/grigata/chta-network-stats/internal/version"
)

const ExplorerURL = "http://chtaexplorer.mooo.com:3002"

//go:embed web/*
var assets embed.FS

type Data struct {
	AppName     string                `json:"app_name"`
	Version     string                `json:"version"`
	Source      string                `json:"source"`
	Height      int64                 `json:"height"`
	ScanSeconds float64               `json:"scan_seconds"`
	GeneratedAt time.Time             `json:"generated_at"`
	ExplorerURL string                `json:"explorer_url"`
	Statistics  statistics.Statistics `json:"statistics"`
	Tags        []models.TagStats     `json:"tags"`
	Blocks      []models.NetworkBlock `json:"blocks"`
}

func NewData(source string, height int64, duration time.Duration, blocks []models.NetworkBlock) Data {
	return Data{
		AppName: version.AppName, Version: version.AppVersion, Source: source,
		Height: height, ScanSeconds: duration.Seconds(), GeneratedAt: time.Now(),
		ExplorerURL: ExplorerURL, Statistics: statistics.Calculate(blocks),
		Tags: statistics.AnalyzeTags(blocks), Blocks: blocks,
	}
}

func Handler(data Data) (http.Handler, error) {
	static, err := fs.Sub(assets, "web")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.Handle("GET /", http.FileServer(http.FS(static)))
	return mux, nil
}

func Serve(ctx context.Context, address string, data Data, open bool) error {
	handler, err := Handler(data)
	if err != nil {
		return fmt.Errorf("prepare dashboard: %w", err)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("start dashboard: %w", err)
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	url := "http://" + listener.Addr().String()
	fmt.Printf("\nDashboard ready: %s\nPress Ctrl+C to stop.\n", url)
	if open {
		go func() { time.Sleep(250 * time.Millisecond); _ = openBrowser(url) }()
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	return command.Start()
}
