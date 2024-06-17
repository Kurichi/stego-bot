package stegobot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

const firebaseURL = "identitytoolkit.googleapis.com"

type (
	SignUpRequest struct {
		ReturnSecureToken bool   `json:"returnSecureToken"`
		DisplayName       string `json:"displayName"`
	}
	SignUpRespnse struct {
		Kind         string `json:"kind"`
		IDToken      string `json:"idToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    string `json:"expiresIn"`
		LocalID      string `json:"localId"`
	}
)

func SignUp(ctx context.Context, apiKey string) (string, error) {
	url := &url.URL{
		Scheme:   "https",
		Host:     firebaseURL,
		Path:     "v1/accounts:signUp",
		RawQuery: "key=" + apiKey,
	}

	req := &SignUpRequest{
		ReturnSecureToken: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer(data)
	resp, err := http.Post(url.String(), "application/json", b)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp SignUpRespnse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.IDToken, nil
}
