// server.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type PingResult struct {
	Subdomain    string `json:"subdomain"`
	IsUp         bool   `json:"is_up"`
	StatusCode   int    `json:"status_code"`
	ResponseTime string `json:"response_time"`
}

type CloudflareResponse struct {
	Success bool `json:"success"`
	Result  []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"result"`
}

// Fungsi menarik data subdomain valid dari Cloudflare (Support Multi-Domain)
func getSubdomainsFromCloudflare() []string {
	var targets []string
	uniqueTargets := make(map[string]bool)

	// Ambil string Zone IDs (dipisahkan koma)
	zoneIDsStr := os.Getenv("CLOUDFLARE_ZONE_IDS")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if zoneIDsStr == "" || apiToken == "" {
		fmt.Println("Peringatan: Kredensial Cloudflare (ZONE_IDS atau API_TOKEN) tidak ditemukan di .env")
		return []string{"https://ryaze.my.id"}
	}

	// Pecah string Zone IDs menjadi array berdasarkan koma
	zoneIDs := strings.Split(zoneIDsStr, ",")
	client := &http.Client{Timeout: 10 * time.Second}

	// Looping untuk setiap Zone ID (Setiap Domain)
	for _, rawZoneID := range zoneIDs {
		zoneID := strings.TrimSpace(rawZoneID)
		if zoneID == "" {
			continue
		}

		url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100", zoneID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Gagal membuat request untuk zone:", zoneID, "-", err)
			continue // Lanjut ke domain berikutnya jika gagal
		}

		req.Header.Add("Authorization", "Bearer "+apiToken)
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Gagal menghubungi Cloudflare untuk zone:", zoneID, "-", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var cfResp CloudflareResponse
		if err := json.Unmarshal(body, &cfResp); err != nil {
			fmt.Println("Gagal parsing JSON untuk zone:", zoneID, "-", err)
			continue
		}

		if !cfResp.Success {
			fmt.Printf("Cloudflare merespons gagal untuk zone %s (cek kredensial)\n", zoneID)
			continue
		}

		// Masukkan hasil record ke antrean target utama
		for _, record := range cfResp.Result {
			if record.Type == "A" || record.Type == "AAAA" || record.Type == "CNAME" {
				if !strings.HasPrefix(record.Name, "*") {
					targetURL := fmt.Sprintf("https://%s", record.Name)

					if !uniqueTargets[targetURL] {
						uniqueTargets[targetURL] = true
						targets = append(targets, targetURL)
					}
				}
			}
		}
	}

	if len(targets) == 0 {
		targets = append(targets, "https://ryaze.my.id")
	}

	return targets
}

func main() {
	godotenv.Load()
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		subdomains := getSubdomainsFromCloudflare()

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

	// Pastikan port hardcoded ke 8010 sesuai perbaikan sebelumnya
	r.Run("127.0.0.1:8010") 
}