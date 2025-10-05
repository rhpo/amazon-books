package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

const maxTry = 5

func Fetch(url string) (string, int, error) {
	return fetchWithRetries(url, 0)
}

func fetchWithRetries(url string, attempt int) (string, int, error) {
	if attempt > maxTry {
		return "", 504, fmt.Errorf("max retries exceeded while fetching: %s", url)
	}

	curlArgs := []string{
		url,
		"--compressed",
		"-H", "User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0",
		"-H", "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"-H", "Accept-Language: en-US,en;q=0.5",
		"-H", "Accept-Encoding: gzip, deflate, br, zstd",
		"-H", "Connection: keep-alive",
		"-H", "Cookie: session-id=258-2014747-6687326; session-id-time=2082787201l; i18n-prefs=EUR; lc-acbfr=fr_FR; csm-hit=tb:08C9C3R57575E60P3TRC+s-N2JDN7ZM15RZ6E8MTV99|1759693231248&t:1759693231248&adb:adblk_no; ubid-acbfr=261-9368196-8565703; session-token=8uzMUfASOHtDkf5a/SKkR3h0djQq2fcXYGsDcO21v7YXP8qqTqRRsQQbmOKfYUrL28Xxr8TwKPNtpUp05jYqC3ucgracn2vu1R8VCoZ0/THyiSIbBn9NuhbFR6NJJoA0aQ6aw6V/NIVqBewPY7JFkvBmpUIAc4d5F+V14uYHTrCdPpgcPrpnBv95pxhTtbfkQ+R8I5Ss+3PNRY+Vq3qwfV/iSVl0IVYUxZwe7wqQsgGwsVkLBE6+fIPHkX1m++NGL29tYSeN/GhANEPKjciwYYO6gBmeXpu1T8cdB3GIPykBT2Gb3iYCBCvBz1JPl8390zxpATm6w8Qt4nC9n7+OgA4+JUasWETN; s_nr=1758565313563-New; s_vnum=2190565313564%26vn%3D1; s_dslv=1758565313564; rxc=AAaOgDXKbjgoGVtO3AQ",
		"-H", "Upgrade-Insecure-Requests: 1",
		"-H", "Sec-Fetch-Dest: document",
		"-H", "Sec-Fetch-Mode: navigate",
		"-H", "Sec-Fetch-Site: cross-site",
		"-H", "Priority: u=0, i",
		"-H", "Pragma: no-cache",
		"-H", "Cache-Control: no-cache",
		"-s",                 // silent mode
		"-w", "%{http_code}", // append HTTP code
	}

	cmd := exec.Command("curl", curlArgs...)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", 500, fmt.Errorf("failed to execute curl: %w", err)
	}

	output := out.String()
	if len(output) < 3 {
		return "", 500, fmt.Errorf("invalid curl output")
	}

	// Extract HTTP status code
	statusStr := output[len(output)-3:]
	body := strings.TrimSpace(output[:len(output)-3])

	status := 0
	fmt.Sscanf(statusStr, "%d", &status)

	if strings.Contains(body, "wait a moment and refresh the page") {
		delay := time.Duration(1+rand.Intn(2)) * time.Second
		fmt.Printf("Request throttled. Waiting %v before retrying...\n", delay)
		time.Sleep(delay)
		return fetchWithRetries(url, attempt+1)
	}

	return body, status, nil
}
