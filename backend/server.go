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

// Struktur data untuk membaca balasan JSON dari Cloudflare
type CloudflareResponse struct {
	Success bool `json:"success"`
	Result  []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"result"`
}

// Fungsi menarik data subdomain valid dari Cloudflare
func getSubdomainsFromCloudflare() []string {
	var targets []string
	uniqueTargets := make(map[string]bool) // Mencegah subdomain ganda (misal punya record A dan AAAA bersamaan)

	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if zoneID == "" || apiToken == "" {
		fmt.Println("Peringatan: Kredensial Cloudflare tidak ditemukan di .env")
		return []string{"https://ryaze.my.id"}
	}

	// Endpoint API Cloudflare untuk mengambil daftar DNS Record
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100", zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Gagal membuat request:", err)
		return []string{"https://ryaze.my.id"}
	}

	// Memasukkan Token API ke Header Authorization
	req.Header.Add("Authorization", "Bearer "+apiToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Gagal menghubungi Cloudflare:", err)
		return []string{"https://ryaze.my.id"}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		fmt.Println("Gagal parsing JSON dari Cloudflare:", err)
		return []string{"https://ryaze.my.id"}
	}

	if !cfResp.Success {
		fmt.Println("Cloudflare merespons API dengan status gagal")
		return []string{"https://ryaze.my.id"}
	}

	// Memasukkan hasil record ke dalam antrean target
	for _, record := range cfResp.Result {
		// Hanya ambil record yang mengarah ke website
		if record.Type == "A" || record.Type == "AAAA" || record.Type == "CNAME" {
			// Abaikan wildcard domain (*)
			if !strings.HasPrefix(record.Name, "*") {
				targetURL := fmt.Sprintf("https://%s", record.Name)

				// Pastikan tidak ada duplikat masuk ke daftar array
				if !uniqueTargets[targetURL] {
					uniqueTargets[targetURL] = true
					targets = append(targets, targetURL)
				}
			}
		}
	}

	// Fallback jika tidak ada record yang ditemukan
	if len(targets) == 0 {
		targets = append(targets, "https://ryaze.my.id")
	}

	return targets
}

func main() {
	// Memuat variabel dari file .env
	godotenv.Load()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
		c.Next()
	})

	r.GET("/api/health", func(c *gin.Context) {
		// Menggunakan fungsi penarik data Cloudflare yang baru
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

	r.Run("127.0.0.1:8010")
}
