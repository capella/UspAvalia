package auth

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HCaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

// VerifyHCaptcha validates an hCaptcha response token with the hCaptcha API
func VerifyHCaptcha(secretKey, responseToken string, r *http.Request) (bool, error) {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	resp, err := http.PostForm("https://hcaptcha.com/siteverify", url.Values{
		"secret":   []string{secretKey},
		"response": []string{responseToken},
		"remoteip": []string{remoteIP},
	})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result HCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return false, fmt.Errorf("hCaptcha verification failed: %s", strings.Join(result.ErrorCodes, ", "))
		}
		return false, fmt.Errorf("hCaptcha verification failed: unknown error")
	}

	return true, nil
}
