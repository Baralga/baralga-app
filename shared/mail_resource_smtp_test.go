package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/matryer/is"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestSendMail(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		return
	}

	is := is.New(t)

	// Setup mail server
	ctx := context.Background()
	cleanupFunc, smtpPort, httpPort, err := setupMailserver(ctx)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := cleanupFunc()
		if err != nil {
			t.Log(err)
		}
	}()

	mailResource := NewSmtpMailResource(
		fmt.Sprintf("localhost:%v", smtpPort),
		"test@baralga.com",
		"test@baralga.com",
		"",
	)

	t.Run("SendMail", func(t *testing.T) {
		// Act
		err := mailResource.SendMail(
			"receiver@baralga.com",
			"Test Subject",
			"Test Body",
		)

		// Assert
		is.NoErr(err)

		totalMessages, err := readTotalMessages(httpPort)
		is.NoErr(err)

		is.Equal(1, totalMessages)
	})
}

func readTotalMessages(httpPort string) (int, error) {
	response, err := http.Get(fmt.Sprintf("http://localhost:%v/api/v2/messages", httpPort))
	if err != nil {
		return 0, err
	}

	var messages MessagesRepresentation
	err = json.NewDecoder(response.Body).Decode(&messages)
	if err != nil {
		return 0, err
	}

	return messages.Total, nil
}

type MessagesRepresentation struct {
	Total int `json:"total"`
}

func setupMailserver(ctx context.Context) (func() error, string, string, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mailhog/mailhog", Tag: "v1.0.1", Env: []string{"listen_addresses = '*'"}},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	apiURI := fmt.Sprintf("http://localhost:%v/api/v2/messages", resource.GetPort("8025/tcp"))

	err = pool.Retry(func() error {
		var err error

		_, err = http.Get(apiURI)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, "", "", err
	}

	return func() error {
		return pool.Purge(resource)
	}, resource.GetPort("1025/tcp"), resource.GetPort("8025/tcp"), nil
}
