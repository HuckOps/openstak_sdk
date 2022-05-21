package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"openstack_sdk/pkg/api_request"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type OpenStackClient struct {
	Token       string
	Expires     time.Time
	NetWork     string
	Image       string
	Placement   string
	Metering    string
	Volumev3    string
	Volumev2    string
	Compute     string
	ObjectStore string
	Alarming    string
	Metric      string
	Identity    string
	User        string
}

type OpenStackConfig struct {
	Clouds struct {
		OpenStack struct {
			Auth struct {
				AuthURL        string `yaml:"auth_url"`
				UserName       string `yaml:"username"`
				ProjectID      string `yaml:"project_id"`
				ProjectName    string `yaml:"project_name"`
				UserDomainName string `yaml:"user_domain_name"`
			} `yaml:"auth"`
			RegionName         string `yaml:"region_name"`
			Interface          string `yaml:"interface"`
			IdentityAPIVersion int    `yaml:"identity_api_version"`
		} `yaml:"openstack"`
	} `yaml:"clouds"`
}

type Auth struct {
	Token struct {
		IssuedAt  time.Time `json:"issued_at"`
		AuditIds  []string  `json:"audit_ids"`
		Methods   []string  `json:"methods"`
		ExpiresAt time.Time `json:"expires_at"`
		User      struct {
			PasswordExpiresAt interface{} `json:"password_expires_at"`
			Domain            struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domain"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
	} `json:"token"`
}

func NewOpenStackClient(config string, password string) (openstackClient OpenStackClient, err error) {
	configfile, err := ioutil.ReadFile(config)
	if err != nil {
		return
	}
	var openstackConfig OpenStackConfig
	if err = yaml.Unmarshal(configfile, &openstackConfig); err != nil {
		return
	}
	client := http.Client{}
	client.Timeout = time.Second * 10
	payloadMap := map[string]interface{}{
		"auth": map[string]interface{}{
			"identity": map[string]interface{}{
				"methods": []string{"password"},
				"password": map[string]interface{}{
					"user": map[string]interface{}{
						"name": openstackConfig.Clouds.OpenStack.Auth.UserName,
						"domain": map[string]interface{}{
							"name": openstackConfig.Clouds.OpenStack.Auth.UserDomainName,
						},
						"password": password,
					},
				},
			},
		},
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return
	}
	payloadReader := bytes.NewReader(payload)
	request, err := http.NewRequest("POST", openstackConfig.Clouds.OpenStack.Auth.AuthURL+"/auth/tokens", payloadReader)
	if err != nil {
		return
	}
	result, err := client.Do(request)
	fmt.Println(err.Error())
	if err != nil || result.StatusCode != 201 {
		panic("Auth Failed")
		return
	}
	response, err := ioutil.ReadAll(result.Body)
	var auth Auth
	if err = json.Unmarshal(response, &auth); err != nil {
		return
	}
	urlSplit := strings.Split(openstackConfig.Clouds.OpenStack.Auth.AuthURL, "://")
	protoco := urlSplit[0]
	host := strings.Split(urlSplit[1], ":")[0]
	openstackClient = OpenStackClient{
		Alarming:    fmt.Sprintf("%s://%s:%s", protoco, host, "8042"),
		Compute:     fmt.Sprintf("%s://%s:%s/v2.1/%s", protoco, host, "8441", openstackConfig.Clouds.OpenStack.Auth.ProjectID),
		Image:       fmt.Sprintf("%s://%s:%s", protoco, host, "9292"),
		Metering:    fmt.Sprintf("%s://%s:%s", protoco, host, "8777"),
		Metric:      fmt.Sprintf("%s://%s:%s", protoco, host, "8041"),
		NetWork:     fmt.Sprintf("%s://%s:%s", protoco, host, "9696"),
		ObjectStore: fmt.Sprintf("%s://%s:%s/v1/AUTH_%s", protoco, host, "8080", openstackConfig.Clouds.OpenStack.Auth.ProjectID),
		Placement:   fmt.Sprintf("%s://%s:%s/placement", protoco, host, "8778"),
		Volumev2:    fmt.Sprintf("%s://%s:%s/v2/%s", protoco, host, "8776", openstackConfig.Clouds.OpenStack.Auth.ProjectID),
		Volumev3:    fmt.Sprintf("%s://%s:%s/v3/%s", protoco, host, "8776", openstackConfig.Clouds.OpenStack.Auth.ProjectID),
		Identity:    openstackConfig.Clouds.OpenStack.Auth.AuthURL,
		Expires:     auth.Token.ExpiresAt,
		Token:       result.Header.Get("X-Subject-Token"),
		User:        auth.Token.User.ID,
	}
	return
}

func (openstackClient OpenStackClient) CheckToken() bool {
	payload := map[string]interface{}{}
	result, err := api_request.SendRequest(api_request.GET, openstackClient.Identity+"/auth/tokens", openstackClient.Token, payload, nil)
	if err != nil || result.StatusCode != 200 {
		panic("Auth Failed")
	}
	return true
}
