package cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"gopkg.in/resty.v1"
)

type Cloud struct {
	Address string
}

type server struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type status struct {
	CpuUtilization int `json:"cpuUtilization"`
}

type AuthSuccess struct {
	/* variables */
}

func (c *Cloud) IsExistServer(name string) bool {
	url, _ := url.Parse(c.Address)
	url.Path = path.Join(url.Path, "check", name)
	resp, err := resty.R().Get(url.String())
	if err != nil {
		return false // TODO need to return err
	}

	if resp.StatusCode() == http.StatusOK {
		return true
	}
	return false
}

func (c *Cloud) IsProhibitedServer(name string) bool {
	url, _ := url.Parse(c.Address)
	url.Path = path.Join(url.Path, "check", name)
	resp, err := resty.R().Get(url.String())
	if err != nil {
		return false // TODO need to return err
	}

	if resp.StatusCode() == http.StatusForbidden {
		return true
	}

	return false
}

func (c *Cloud) GetUUID(name string) (string, error) {
	servers := []server{}
	url, _ := url.Parse(c.Address)
	url.Path = path.Join(url.Path, "servers")

	resp, err := resty.R().Get(url.String())
	if err != nil {
		return "", err
	}
	json.Unmarshal(resp.Body(), &servers)
	for _, server := range servers {
		if server.Name == name {
			return server.ID, nil
		}
	}
	return "", fmt.Errorf("server not found")
}

func (c *Cloud) getStatusByUUID(uuid string) (string, int, error) {
	status := status{}
	url, _ := url.Parse(c.Address)
	url.Path = path.Join(url.Path, "servers", uuid, "status")
	resp, err := resty.R().Get(url.String())
	if err != nil {
		return "", -1, err
	}
	json.Unmarshal(resp.Body(), &status)
	return uuid, status.CpuUtilization, nil
}

func (c *Cloud) GetStatus(name string) (string, int, error) {
	uuid, err := c.GetUUID(name)
	if err != nil {
		return "", -1, nil
	}

	return c.getStatusByUUID(uuid)
}

func (c *Cloud) CreateServer(name string) error {
	server := server{Name: name}
	body, err := json.Marshal(server)
	if err != nil {
		return err
	}

	url, _ := url.Parse(c.Address)
	url.Path = path.Join(url.Path, "servers")

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&AuthSuccess{}).
		Post(url.String())

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusCreated {
		return nil
	}
	return fmt.Errorf("failed to create %s Status code %v", name, resp.StatusCode())
}

func (c *Cloud) DeleteServer(name string) error {
	uuid, err := c.GetUUID(name)
	if err != nil {
		return err
	}
	url, _ := url.Parse(c.Address)
	url.Path = path.Join(url.Path, "servers", uuid)

	resp, err := resty.R().
		Delete(url.String())
	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("failed to delete %s", name)
}
