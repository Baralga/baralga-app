package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/matryer/is"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSendMail(t *testing.T) {
	// skip in short mode
	if testing.Short() {
		return
	}

	is := is.New(t)

	// Setup mail server
	ctx := context.Background()
	mailContainer, host, smtpPort, httpPort, err := setupMailserver(ctx)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := mailContainer.Terminate(ctx)
		if err != nil {
			t.Log(err)
		}
	}()

	mailResource := NewSmtpMailResource(
		fmt.Sprintf("%v:%v", host, smtpPort),
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

		totalMessages, err := readTotalMessages(host, httpPort)
		is.NoErr(err)

		is.Equal(1, totalMessages)
	})
}

func readTotalMessages(host, httpPort string) (int, error) {
	response, err := http.Get(fmt.Sprintf("http://%v:%v/api/v2/messages", host, httpPort))
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

func setupMailserver(ctx context.Context) (testcontainers.Container, string, string, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog:v1.0.1",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForListeningPort("1025/tcp"),
		Env:          map[string]string{},
	}
	mailContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", "", "", err
	}
	smtpPort, err := mailContainer.MappedPort(ctx, "1025")
	if err != nil {
		return nil, "", "", "", err
	}
	httpPort, err := mailContainer.MappedPort(ctx, "8025")
	if err != nil {
		return nil, "", "", "", err
	}
	host, err := mailContainer.Host(ctx)
	if err != nil {
		return nil, "", "", "", err
	}

	return mailContainer, host, smtpPort.Port(), httpPort.Port(), nil
}
