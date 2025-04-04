package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	url := query.Get("url")
	referer := query.Get("referer")

	if url == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Write(body)
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	url := query.Get("url")
	query.Del("url")

	client := &http.Client{}

	if len(query) > 0 {
		url += "?" + query.Encode()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create request for %s: %v", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error occurred while requesting %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body from %s: %v", url, err)
		}

		content := string(body)
		if _, ok := query["proxy"]; ok {
			content = replaceImgWithTemplate(content, r.Host, query.Get("referer"))
		}

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		contentLength := len(content)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))

		w.Write([]byte(content))
		return
	} else {
		log.Printf("Non-200 status code (%d) from %s", resp.StatusCode, url)
	}

	http.Error(w, "No website instance returned a 200 status code", http.StatusBadGateway)
}

func replaceImgWithTemplate(content, host, referer string) string {
	pattern := `<img\s+src="([^"]+)"`
	pattern2 := `&lt;img\s+src=\&quot;(.+?)\&quot;`
	pattern3 := `&lt;img\s+src=\&#34;(.+?)\&#34;`

	proxyImgTemplate := func(url string) string {
		return fmt.Sprintf("http://%s/image?url=%s&referer=%s", host, url, referer)
	}

	re := regexp.MustCompile(pattern)
	re2 := regexp.MustCompile(pattern2)
	re3 := regexp.MustCompile(pattern3)

	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		newURL := proxyImgTemplate(match[1])
		content = strings.Replace(content, match[0], fmt.Sprintf(`<img src="%s"`, newURL), 1)
	}

	matches2 := re2.FindAllStringSubmatch(content, -1)
	for _, match := range matches2 {
		newURL := strings.ReplaceAll(proxyImgTemplate(match[1]), "&", "&amp;")
		content = strings.Replace(content, match[0], fmt.Sprintf(`&lt;img src=&quot;%s&quot;`, newURL), 1)
	}

	matches3 := re3.FindAllStringSubmatch(content, -1)
	for _, match := range matches3 {
		newURL := strings.ReplaceAll(proxyImgTemplate(match[1]), "&", "&#38;")
		content = strings.Replace(content, match[0], fmt.Sprintf(`&lt;img src=&#34;%s&#34;`, newURL), 1)
	}

	return content
}

func main() {
	http.HandleFunc("/image", proxyHandler)
	http.HandleFunc("/rss", rssHandler)

	port := 8080
	log.Printf("Starting server on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
