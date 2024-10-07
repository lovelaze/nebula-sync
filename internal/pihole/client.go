package pihole

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

var (
	userAgent  = fmt.Sprintf("nebula-sync/%s", version.Version)
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

func NewClient(piHole model.PiHole) Client {
	logger := log.With().Str("client", piHole.Url.String()).Logger()
	return &client{
		piHole: piHole,
		logger: &logger,
	}
}

type Client interface {
	Authenticate() error
	DeleteSession() error
	GetVersion() (*model.VersionResponse, error)
	GetTeleporter() ([]byte, error)
	PostTeleporter(payload []byte, teleporterRequest *model.PostTeleporterRequest) error
	GetConfig() (configResponse *model.ConfigResponse, err error)
	PatchConfig(patchRequest *model.PatchConfigRequest) error
	String() string
	ApiPath(target string) string
}

type client struct {
	piHole model.PiHole
	auth   auth
	logger *zerolog.Logger
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

	if a.sid == "" {
		return errors.New("no sid found")
	}

	if a.validity <= 0 {
		return errors.New("expired sid found")
	}

	return nil
}

func (client *client) Authenticate() error {
	client.logger.Debug().Msg("Authenticate")
	authResponse := model.AuthResponse{}

	reqBytes, err := json.Marshal(model.AuthRequest{Password: client.piHole.Password})
	if err != nil {
		return client.wrapError(err, nil)
	}

	req, err := http.NewRequest("POST", client.ApiPath("/auth"), bytes.NewReader(reqBytes))

	if err != nil {
		return client.wrapError(err, req)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
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

	req, err := http.NewRequest("DELETE", client.ApiPath("auth"), nil)
	if err != nil {
		return client.wrapError(err, req)
	}

	req.Header.Set("sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)

	if err != nil {
		return client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return client.wrapError(err, req)
}

func (client *client) GetVersion() (*model.VersionResponse, error) {
	client.logger.Debug().Msg("Get version")
	versionResponse := model.VersionResponse{}
	if err := client.auth.verify(); err != nil {
		return &versionResponse, client.wrapError(err, nil)
	}

	req, err := http.NewRequest("GET", client.ApiPath("info/version"), nil)
	if err != nil {
		return &versionResponse, client.wrapError(err, req)
	}
	req.Header.Set("sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return &versionResponse, client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
		return &versionResponse, client.wrapError(err, req)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return &versionResponse, client.wrapError(err, req)
	}

	err = json.Unmarshal(body, &versionResponse)

	return &versionResponse, client.wrapError(err, req)
}

func (client *client) GetTeleporter() ([]byte, error) {
	client.logger.Debug().Msg("Get teleporter")
	if err := client.auth.verify(); err != nil {
		return nil, client.wrapError(err, nil)
	}
	req, err := http.NewRequest("GET", client.ApiPath("teleporter"), nil)
	if err != nil {
		return nil, client.wrapError(err, req)
	}
	req.Header.Set("sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
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

	req, err := http.NewRequest("POST", client.ApiPath("teleporter"), &requestBody)
	if err != nil {
		return client.wrapError(err, req)
	}
	req.Header.Set("sid", client.auth.sid)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return nil
}

func (client *client) GetConfig() (configResponse *model.ConfigResponse, err error) {
	client.logger.Debug().Msg("Get config")
	if err := client.auth.verify(); err != nil {
		return configResponse, client.wrapError(err, nil)
	}

	req, err := http.NewRequest("GET", client.ApiPath("config"), nil)
	if err != nil {
		return configResponse, client.wrapError(err, req)
	}
	req.Header.Set("sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return configResponse, client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
		return configResponse, client.wrapError(err, req)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return configResponse, client.wrapError(err, req)
	}

	if err := json.Unmarshal(body, &configResponse); err != nil {
		return configResponse, client.wrapError(err, req)
	}

	return configResponse, client.wrapError(err, req)
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

	req, err := http.NewRequest("PATCH", client.ApiPath("config"), bytes.NewReader(reqBytes))
	if err != nil {
		return client.wrapError(err, req)
	}
	req.Header.Set("sid", client.auth.sid)
	req.Header.Set("User-Agent", userAgent)

	response, err := httpClient.Do(req)
	if err != nil {
		return client.wrapError(err, req)
	}

	if err := successfulHttpStatus(response.StatusCode); err != nil {
		return client.wrapError(err, req)
	}

	return client.wrapError(err, req)
}

func (client *client) String() string {
	return client.piHole.Url.String()
}

func (client *client) ApiPath(target string) string {
	return client.piHole.Url.JoinPath("api", target).String()
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

func successfulHttpStatus(statusCode int) error {
	if statusCode >= 200 && statusCode <= 299 {
		return nil
	}

	return fmt.Errorf("unexpected status code: %d", statusCode)
}
