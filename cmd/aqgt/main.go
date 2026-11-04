package main

import (
	"aquadelic-gt-integrated-patcher/engine"
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

//go:embed index.html
var pages embed.FS

type request struct {
	Game string `json:"game"`
	Mode string `json:"mode"`
}
type response struct {
	OK         bool               `json:"ok"`
	Error      string             `json:"error,omitempty"`
	Message    string             `json:"message,omitempty"`
	Platform   string             `json:"platform,omitempty"`
	Candidates []engine.Candidate `json:"candidates,omitempty"`
	Status     *engine.Status     `json:"status,omitempty"`
	Path       string             `json:"path,omitempty"`
}

func main() {
	action := flag.String("action", "gui", "gui | status | install | restore")
	game := flag.String("game", "", "Game installation directory")
	mode := flag.String("controls", "none", "none | analog | step")
	yes := flag.Bool("yes", false, "Confirm the game and launcher are closed and allow file changes")
	flag.Parse()
	if *action != "gui" {
		if *game == "" {
			fmt.Fprintln(os.Stderr, "Provide --game /path/to/Aquadelic GT")
			os.Exit(2)
		}
		switch *action {
		case "status":
			b, _ := json.MarshalIndent(engine.Check(*game), "", "  ")
			fmt.Println(string(b))
		case "install", "restore":
			if !*yes {
				fmt.Fprintln(os.Stderr, "STOP: --yes is required")
				os.Exit(2)
			}
			var message string
			var err error
			if *action == "install" {
				message, err = engine.Install(*game, *mode)
			} else {
				message, err = engine.Restore(*game)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "STOP:", err)
				os.Exit(1)
			}
			fmt.Println(message)
		default:
			fmt.Fprintln(os.Stderr, "Unknown action")
			os.Exit(2)
		}
		return
	}
	if err := gui(); err != nil {
		fmt.Fprintln(os.Stderr, "Aquadelic Patcher:", err)
		os.Exit(1)
	}
}
func send(w http.ResponseWriter, code int, r response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(r)
}
func gui() error {
	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)
	t, err := template.ParseFS(pages, "index.html")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	// Serialise potentially destructive operations from multiple browser tabs.
	var writes sync.Mutex
	var server *http.Server
	security := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("X-Frame-Options", "DENY")
			if r.Method != "POST" || r.Header.Get("X-AQGT-Token") != token {
				send(w, 403, response{Error: "Not authorized"})
				return
			}
			if r.ContentLength > 64*1024 {
				send(w, 413, response{Error: "Request too large"})
				return
			}
			origin := r.Header.Get("Origin")
			if origin != "" && origin != "http://"+r.Host {
				send(w, 403, response{Error: "Origin not allowed"})
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
			next(w, r)
		}
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
		_ = t.Execute(w, struct{ Token string }{token})
	})
	api := func(fn func(request) (response, error)) http.HandlerFunc {
		return security(func(w http.ResponseWriter, r *http.Request) {
			var req request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				send(w, 400, response{Error: "Invalid request data"})
				return
			}
			answer, err := fn(req)
			if err != nil {
				send(w, 400, response{Error: err.Error()})
				return
			}
			answer.OK = true
			send(w, 200, answer)
		})
	}
	mux.HandleFunc("/api/discover", api(func(_ request) (response, error) {
		return response{Candidates: engine.Discover(), Platform: engine.Platform()}, nil
	}))
	mux.HandleFunc("/api/check", api(func(r request) (response, error) {
		st := engine.Check(r.Game)
		if st.Active || st.Ready {
			engine.RememberGame(st.Game)
		}
		return response{Status: &st}, nil
	}))
	mux.HandleFunc("/api/install", api(func(r request) (response, error) {
		writes.Lock()
		defer writes.Unlock()
		s, e := engine.Install(r.Game, r.Mode)
		if e == nil {
			engine.RememberGame(r.Game)
		}
		return response{Message: s}, e
	}))
	mux.HandleFunc("/api/restore", api(func(r request) (response, error) {
		writes.Lock()
		defer writes.Unlock()
		s, e := engine.Restore(r.Game)
		if e == nil {
			engine.RememberGame(r.Game)
		}
		return response{Message: s}, e
	}))
	mux.HandleFunc("/api/pick", api(func(_ request) (response, error) { return response{Path: pickFolder()}, nil }))
	mux.HandleFunc("/api/exit", security(func(w http.ResponseWriter, r *http.Request) {
		send(w, 200, response{OK: true})
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		}()
	}))
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	server = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	url := "http://" + listener.Addr().String() + "/"
	fmt.Println("Aquadelic GT Compatibility Patcher 1.1.0 RC5 —", url)
	go func() {
		time.Sleep(300 * time.Millisecond)
		var open *exec.Cmd
		if runtime.GOOS == "windows" {
			open = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url)
		} else {
			open = exec.Command("xdg-open", url)
		}
		if e := open.Start(); e != nil {
			fmt.Println("Open in your browser:", url)
		} else {
			go func() { _ = open.Wait() }()
		}
	}()
	err = server.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
