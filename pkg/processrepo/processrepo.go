package processrepo

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/model/processmodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

func New(config configuration.Config) controller.ProcessRepo {
	return &Repository{
		config: config,
	}
}

type Repository struct {
	config configuration.Config
}

func (this *Repository) GetProcessModel(token string, id string) (result processmodel.ProcessModel, err error, errCode int) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.ProcessRepoUrl+"/processes/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	return result, nil, 200
}
