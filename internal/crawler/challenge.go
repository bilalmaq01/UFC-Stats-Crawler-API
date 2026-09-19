package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var (
	nonceRe = regexp.MustCompile(`nonce="([0-9a-f]+)"`)
	diffRe  = regexp.MustCompile(`target=new Array\((\d+)\+1\)`)
)

// parseChallenge pulls the proof-of-work nonce and difficulty (number of leading
// hex zeros the hash must have) out of a ufcstats "Loading…" challenge page.
// ok is false when the page isn't a challenge (i.e. we already have real HTML).
func parseChallenge(html string) (nonce string, zeros int, ok bool) {
	nm := nonceRe.FindStringSubmatch(html)
	dm := diffRe.FindStringSubmatch(html)
	if nm == nil || dm == nil {
		return "", 0, false
	}
	zeros, _ = strconv.Atoi(dm[1])
	return nm[1], zeros, true
}

// Bootstrap runs the "errand" once: GET the probe page, and if it's a challenge,
// solve the proof-of-work and POST the answer to /__c. The resulting cookie lands
// in jar, so any client/collector sharing jar is waved straight through after.
// No-op if the probe isn't a challenge (cookie already valid, or none needed).
func Bootstrap(jar http.CookieJar, probeURL string) error {
	client := &http.Client{Jar: jar}

	body, err := fetch(client, http.MethodGet, probeURL, "")
	if err != nil {
		return err
	}
	nonce, zeros, ok := parseChallenge(body)
	if !ok {
		return nil
	}

	u, err := url.Parse(probeURL)
	if err != nil {
		return err
	}
	form := "nonce=" + nonce + "&n=" + strconv.Itoa(solvePoW(nonce, zeros))
	_, err = fetch(client, http.MethodPost, u.Scheme+"://"+u.Host+"/__c", form)
	return err
}

// fetch does one HTTP request with the browser User-Agent, returning the body.
func fetch(client *http.Client, method, rawURL, body string) (string, error) {
	req, err := http.NewRequest(method, rawURL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// solvePoW brute-forces the smallest n where sha256("<nonce>:<n>") begins with
// `zeros` hex zeros. Cheap: difficulty 2 averages ~256 tries.
func solvePoW(nonce string, zeros int) int {
	prefix := strings.Repeat("0", zeros)
	for n := 0; ; n++ {
		sum := sha256.Sum256([]byte(nonce + ":" + strconv.Itoa(n)))
		if strings.HasPrefix(hex.EncodeToString(sum[:]), prefix) {
			return n
		}
	}
}
