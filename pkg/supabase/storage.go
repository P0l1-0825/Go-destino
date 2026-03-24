// Package supabase provides helpers for Supabase Storage and Realtime.
package supabase

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// StorageBucket represents an allowed Supabase Storage bucket.
type StorageBucket string

const (
	BucketTickets     StorageBucket = "tickets"
	BucketQRCodes     StorageBucket = "qr-codes"
	BucketDriverDocs  StorageBucket = "driver-docs"
	BucketKioskDiag   StorageBucket = "kiosk-diagnostics"
)

// StorageClient uploads files to Supabase Storage via REST API.
type StorageClient struct {
	baseURL    string // e.g. https://PROJECT.supabase.co
	serviceKey string // service_role key (bypasses RLS)
	httpClient *http.Client
}

// NewStorageClient creates a client from env vars.
// Reads SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY.
func NewStorageClient() (*StorageClient, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY required")
	}
	return &StorageClient{
		baseURL:    url,
		serviceKey: key,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Upload stores a file in Supabase Storage.
// path format: "tenant_id/entity_id/filename.ext"
func (s *StorageClient) Upload(bucket StorageBucket, path string, data []byte, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, bucket, path)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("uploading to storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("storage upload failed (%d): %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, bucket, path)
	return publicURL, nil
}

// PublicURL returns the public URL for a stored object.
func (s *StorageClient) PublicURL(bucket StorageBucket, path string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, bucket, path)
}

// Delete removes a file from storage.
func (s *StorageClient) Delete(bucket StorageBucket, path string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, bucket, path)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("storage delete failed (%d)", resp.StatusCode)
	}
	return nil
}
