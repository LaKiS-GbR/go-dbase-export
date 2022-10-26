package job

import (
	"crypto/rand"
	"encoding/base64"
)

type Manager struct {
	Jobs []*Job
}

func NewManager() *Manager {
	return &Manager{
		Jobs: make([]*Job, 0),
	}
}

func (m *Manager) AddJob(j *Job) error {
	// Generate a unique ID for the job
	id, err := getRandomBytesBase64(32)
	if err != nil {
		return err
	}

	j.ID = id
	m.Jobs = append(m.Jobs, j)
	return nil
}

func (m *Manager) GetJobs() []*Job {
	return m.Jobs
}

func getRandomBytes(size uint) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GetRandomBytesBase64 Returns the Base64 Encoded Equivalent of
// calling GetRandomBytes.
func getRandomBytesBase64(size uint) (string, error) {
	bytes, err := getRandomBytes(size)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
