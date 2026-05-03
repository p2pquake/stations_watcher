package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/p2pquake/stations_watcher/internal/storage"
)

type CreateIssueParams struct {
	Org            string
	Repo           string
	AppID          string
	PemBucket      string
	PemKey         string
	InstallationID string
	Diff           string
}

func CreateIssue(ctx context.Context, p CreateIssueParams) error {
	pem, err := storage.GetObjectBytes(ctx, p.PemBucket, p.PemKey)
	if err != nil {
		return fmt.Errorf("get pem: %w", err)
	}
	jwtToken, err := generateAppJWT(pem, p.AppID)
	if err != nil {
		return fmt.Errorf("generate jwt: %w", err)
	}
	accessToken, err := getAccessToken(ctx, jwtToken, p.InstallationID)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	body := fmt.Sprintf(
		"The list of seismic intensity stations has been updated.\n"+
			"See: https://www.jma.go.jp/jma/kishou/know/jishin/intens-st/stations.json\n\n"+
			"```diff\n%s\n```\n\n"+
			"Created by stations_watcher (https://github.com/p2pquake/stations_watcher)",
		p.Diff,
	)

	payload, err := json.Marshal(map[string]string{
		"title": "Update seismic intensity stations",
		"body":  body,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", p.Org, p.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("User-Agent", "stations_watcher")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	return nil
}

type accessTokenResponse struct {
	Token string `json:"token"`
}

func getAccessToken(ctx context.Context, jwtToken, installationID string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "stations_watcher")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var tok accessTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("decode access token: %w (body: %s)", err, string(body))
	}
	if tok.Token == "" {
		return "", fmt.Errorf("empty access token (body: %s)", string(body))
	}
	return tok.Token, nil
}

func generateAppJWT(privateKeyPEM []byte, appID string) (string, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": appID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}
