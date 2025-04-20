package pihole

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/version"
)

var userAgent = fmt.Sprintf("nebula-sync/%s", version.Version)

type Client interface {
	PostAuth() error
	DeleteSession() error
	GetTeleporter() ([]byte, error)
	PostTeleporter(payload []byte, teleporterRequest *model.PostTeleporterRequest) error
	GetConfig() (configResponse *model.ConfigResponse, err error)
	PatchConfig(patchRequest *model.PatchConfigRequest) error
	PostRunGravity() error
	String() string
	APIPath(target string) string
}

func NewClient(piHole model.PiHole, httpClient *http.Client) Client {
	logger := log.With().Str("client", piHole.URL.String()).Logger()
	return &client{
		piHole:     piHole,
		logger:     &logger,
		httpClient: httpClient,
	}
}

type client struct {
	piHole     model.PiHole
	auth       auth
	logger     *zerolog.Logger
	httpClient *http.Client
}

type auth struct {
	sid      string
	csrf     string
	validity int
	valid    bool
}

func (a *auth) verify() error {
	if !a.valid {
		return errors.New("invalid sid found")
	}

	return nil
}

func (client *client) PostAuth() error {
	client.logger.Debug().Msg("PostAuth")
	authResponse := model.AuthResponse{}

	reqBytes, err := json.Marshal(model.AuthRequest{Password: client.piHole.Password})
	if err != nil {
		return client.wrapError(err, nil)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		client.APIPath("/auth"),
		bytes.NewReader(reqBytes),
	)
	if err != nil {
		return client.wrapError(err, req)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return client.wrapError(err, req)
	}

	if err = json.Unmarshal(body, &authResponse); err != nil {
		return client.wrapError(err, req)
	}

	client.auth = auth{
		sid:      authResponse.Session.Sid,
		csrf:     authResponse.Session.Csrf,
		validity: authResponse.Session.Validity,
		valid:    authResponse.Session.Valid,
	}

	return client.auth.verify()
}

func (client *client) DeleteSession() error {
	client.logger.Debug().Msg("Delete session")
	if err := client.auth.verify(); err != nil {
		return client.wrapError(err, nil)
	}

	if client.auth.sid == "" {
		log.Debug().Msg("Trying to delete empty session")
		return nil
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, client.APIPath("auth"), nil)
	if err != nil {
		return client.wrapError(err, req)
	}

	req.Header.Set("Sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return client.wrapError(err, req)
}

func (client *client) GetTeleporter() ([]byte, error) {
	client.logger.Debug().Msg("Get teleporter")
	if err := client.auth.verify(); err != nil {
		return nil, client.wrapError(err, nil)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, client.APIPath("teleporter"), nil)
	if err != nil {
		return nil, client.wrapError(err, req)
	}
	req.Header.Set("Sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return nil, client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return nil, client.wrapError(err, req)
	}

	body, err := io.ReadAll(response.Body)
	return body, client.wrapError(err, req)
}

func (client *client) PostTeleporter(payload []byte, teleporterRequest *model.PostTeleporterRequest) error {
	client.logger.Debug().Any("payload", teleporterRequest).Msg("Post teleporter")

	if err := client.auth.verify(); err != nil {
		return client.wrapError(err, nil)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	fileWriter, _ := writer.CreateFormFile("file", "config.zip")
	if _, err := io.Copy(fileWriter, bytes.NewReader(payload)); err != nil {
		return client.wrapError(err, nil)
	}

	if teleporterRequest != nil {
		jsonData, err := json.Marshal(teleporterRequest)
		if err != nil {
			return client.wrapError(err, nil)
		}
		if err = writer.WriteField("import", string(jsonData)); err != nil {
			return client.wrapError(err, nil)
		}
	}

	if err := writer.Close(); err != nil {
		return client.wrapError(err, nil)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, client.APIPath("teleporter"), &requestBody)
	if err != nil {
		return client.wrapError(err, req)
	}
	req.Header.Set("Sid", client.auth.sid)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return nil
}

func (client *client) GetConfig() (*model.ConfigResponse, error) {
	var configResponse model.ConfigResponse
	client.logger.Debug().Msg("Get config")
	if err := client.auth.verify(); err != nil {
		return &configResponse, client.wrapError(err, nil)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, client.APIPath("config"), nil)
	if err != nil {
		return &configResponse, client.wrapError(err, req)
	}
	req.Header.Set("Sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return &configResponse, client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return &configResponse, client.wrapError(err, req)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return &configResponse, client.wrapError(err, req)
	}

	if err := json.Unmarshal(body, &configResponse); err != nil {
		return &configResponse, client.wrapError(err, req)
	}

	return &configResponse, client.wrapError(err, req)
}

func (client *client) PatchConfig(patchRequest *model.PatchConfigRequest) error {
	client.logger.Debug().Any("payload", patchRequest).Msgf("Patch config")
	if err := client.auth.verify(); err != nil {
		return client.wrapError(err, nil)
	}

	reqBytes, err := json.Marshal(patchRequest)
	if err != nil {
		return client.wrapError(err, nil)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPatch,
		client.APIPath("config"),
		bytes.NewReader(reqBytes),
	)
	if err != nil {
		return client.wrapError(err, req)
	}
	req.Header.Set("Sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return client.wrapError(err, req)
}

func (client *client) PostRunGravity() error {
	client.logger.Debug().Msg("Post run gravity")
	if err := client.auth.verify(); err != nil {
		return client.wrapError(err, nil)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, client.APIPath("action/gravity"), nil)
	if err != nil {
		return client.wrapError(err, req)
	}
	req.Header.Set("Sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}
	defer response.Body.Close()

	if err := successfulHTTPStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return err
}

func (client *client) String() string {
	return client.piHole.URL.String()
}

func (client *client) APIPath(target string) string {
	return client.piHole.URL.JoinPath("api", target).String()
}

func (client *client) wrapError(err error, req *http.Request) error {
	if err != nil {
		if req != nil {
			return fmt.Errorf("%s: %w", req.URL.String(), err)
		}
		return fmt.Errorf("%s: %w", client.String(), err)
	}
	return nil
}

func successfulHTTPStatus(statusCode int) error {
	if statusCode >= 200 && statusCode <= 299 {
		return nil
	}

	return fmt.Errorf("unexpected status code: %d", statusCode)
}
