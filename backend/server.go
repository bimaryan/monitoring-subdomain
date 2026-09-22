// server.go
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type PingResult struct {
	Subdomain    string `json:"subdomain"`
	IsUp         bool   `json:"is_up"`
	StatusCode   int    `json:"status_code"`
	ResponseTime string `json:"response_time"`
}

// Fungsi untuk membaca folder website 1Panel otomatis
func getDynamicSubdomains() []string {
	var targets []string
	
	// Sesuaikan dengan letak folder sites 1Panel kamu. 
	rootDir := "/opt/1panel/apps/openresty/openresty/www/sites/" 

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		fmt.Println("Gagal membaca direktori:", err)
		// Fallback data statis jika folder tidak ditemukan
		return []string{
			"https://ryaze.my.id",
			"https://bimz.my.id",
			"https://ryz.my.id",
			"https://ryaze.cloud",
			"https://safetalkai.my.id",
		}
	}

	for _, entry := range entries {
		// Pastikan yang dibaca hanya sebuah folder
		if entry.IsDir() {
			name := entry.Name()
			
			// Filter: Abaikan folder sistem 1Panel atau folder tersembunyi
			if name == "log" || name == "default_error_pages" || name == "hosting_clients" || strings.HasPrefix(name, ".") {
				continue
			}

			// Filter: Pastikan nama folder mengandung titik (.) sebagai penanda format domain
			// Ini mengabaikan folder wildcard seperti "*.ryaze.my.id" jika tidak bisa diping langsung
			if strings.Contains(name, ".") && !strings.HasPrefix(name, "*") {
				// Nama folder di 1Panel sama dengan nama domain, langsung gabungkan dengan protokol
				targetURL := fmt.Sprintf("https://%s", name)
				targets = append(targets, targetURL)
			}
		}
	}

	return targets
}

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		// Panggil fungsi pencari subdomain otomatis di sini
		subdomains := getDynamicSubdomains()

		var results []PingResult
		var wg sync.WaitGroup
		var mu sync.Mutex

		client := http.Client{
			Timeout: 5 * time.Second,
		}

		for _, url := range subdomains {
			wg.Add(1)

			go func(targetURL string) {
				defer wg.Done()

				start := time.Now()
				resp, err := client.Get(targetURL)
				duration := time.Since(start)

				result := PingResult{
					Subdomain:    targetURL,
					ResponseTime: duration.Round(time.Millisecond).String(),
				}

				if err != nil {
					result.IsUp = false
					result.StatusCode = 0
				} else {
					defer resp.Body.Close()
					result.StatusCode = resp.StatusCode
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						result.IsUp = true
					} else {
						result.IsUp = false
					}
				}

				mu.Lock()
				results = append(results, result)
				mu.Unlock()

			}(url)
		}

		wg.Wait()

		c.JSON(200, gin.H{
			"status": "success",
			"data":   results,
		})
	})

	// Menggunakan 127.0.0.1 agar lebih aman, hanya diakses lewat Next.js proxy
	r.Run("127.0.0.1:8010")
}
