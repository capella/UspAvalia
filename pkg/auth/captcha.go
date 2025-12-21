package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type HCaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

// VerifyHCaptcha validates an hCaptcha response token with the hCaptcha API
func VerifyHCaptcha(secretKey, responseToken, remoteIP string) (bool, error) {
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

	return result.Success, nil
}
