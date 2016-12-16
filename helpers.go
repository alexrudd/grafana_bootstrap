package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type config struct {
	Datasources   map[string]DataSourceDTO `json:"datasources" yaml:"datasources"`
	Dashboards    map[string]dashboard     `json:"dashboards" yaml:"dashboards"`
	Organisations []organisation           `json:"organisations" yaml:"organisations"`
}

// DataSourceDTO https://github.com/grafana/grafana/blob/master/pkg/api/dtos/models.go
type DataSourceDTO struct {
	Access            string `json:"access" yaml:"access"`
	BasicAuth         bool   `json:"basicAuth" yaml:"basicAuth"`
	BasicAuthPassword string `json:"basicAuthPassword" yaml:"basicAuthPassword"`
	BasicAuthUser     string `json:"basicAuthUser" yaml:"basicAuthUser"`
	Database          string `json:"database" yaml:"database"`
	ID                int64  `json:"id,omitempty" yaml:"id,omitempty"`
	IsDefault         bool   `json:"isDefault" yaml:"isDefault"`
	JSONData          struct {
		EsVersion int    `json:"esVersion" yaml:"esVersion"`
		Interval  string `json:"interval" yaml:"interval"`
		TimeField string `json:"timeField" yaml:"timeField"`
	} `json:"jsonData,omitempty" yaml:"jsonData,omitempty"`
	Name            string `json:"name" yaml:"name"`
	OrgID           int64  `json:"orgId" yaml:"orgId"`
	Password        string `json:"password" yaml:"password"`
	Type            string `json:"type" yaml:"type"`
	TypeLogoURL     string `json:"typeLogoUrl" yaml:"typeLogoUrl"`
	URL             string `json:"url" yaml:"url"`
	User            string `json:"user" yaml:"user"`
	WithCredentials bool   `json:"withCredentials" yaml:"withCredentials"`
}

type dashboard struct {
	Name string `json:"name" yaml:"name"`
	File string `json:"file" yaml:"file"`
}

// DashboardDTO for creating/updating dashboards
type DashboardDTO struct {
	Overwrite bool            `json:"overwrite"`
	Dashboard json.RawMessage `json:"dashboard"`
}

type organisation struct {
	ID            int64             `json:"id,omitempty"`
	Name          string            `json:"name" yaml:"name"`
	APIKey        string            `json:"apiKey" yaml:"apiKey"`
	Datasources   []string          `json:"datasources" yaml:"datasources"`
	Dashboards    []string          `json:"dashboards" yaml:"dashboards"`
	DashboardVars map[string]string `json:"dashboardVars" yaml:"dashboardVars"`
}

type anyID struct {
	ID int64 `json:"id"`
}

func doesOrgExist(ep, name string) (*anyID, error) {
	resp, err := http.Get(ep + "api/orgs/name/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	gOrg := anyID{}
	err = json.Unmarshal(body, &gOrg)
	if err != nil {
		return nil, err
	}
	return &gOrg, nil
}

func createOrg(ep, name string) (*anyID, error) {
	resp, err := http.Post(ep+"api/orgs", "application/json", strings.NewReader(`{"name":"`+name+`"}`))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	gOrg := anyID{}
	err = json.Unmarshal(body, &gOrg)
	if err != nil {
		return nil, err
	}
	return &gOrg, nil
}

func createUpdateDatasource(ep string, org organisation, dsName string, ds DataSourceDTO) error {
	// check api key exists
	//log.Debug(ds.JSONData)
	if org.APIKey == "" {
		return errors.New("Missing API key for Org: " + strconv.Itoa(int(org.ID)))
	}
	// get existing if it exists
	req, err := http.NewRequest("GET", ep+"api/datasources/name/"+dsName, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+org.APIKey)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debugf("Does org %v datasource \"%s\" exist? %s", org.ID, dsName, resp.Status)

	if resp.StatusCode == 200 {
		// Datasource exists!
		log.Debugf("Datasource \"%s\" exists", dsName)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		gDs := anyID{}
		err = json.Unmarshal(body, &gDs)
		if err != nil {
			return err
		}
		ds.ID = gDs.ID
		ds.Name = dsName
		ds.OrgID = org.ID
		data, err := json.Marshal(ds)
		if err != nil {
			return err
		}
		req, err = http.NewRequest("PUT", ep+"api/datasources/"+strconv.Itoa(int(ds.ID)), strings.NewReader(string(data)))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+org.APIKey)
		resp, err = (&http.Client{}).Do(req)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return nil
		}
		return errors.New(resp.Status)
	}
	// Datasource doesn't exist
	log.Debugf("Datasource \"%s\" doesn't exists", dsName)
	ds.Name = dsName
	ds.OrgID = org.ID
	data, err := json.Marshal(ds)
	if err != nil {
		return err
	}
	log.Debugf("Turned into JSON: %s", string(data))
	req, err = http.NewRequest("POST", ep+"api/datasources/", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+org.APIKey)
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	log.Debugf("Posted \"%s\"", dsName)
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	return errors.New(resp.Status)
}

func createUpdateDashboard(ep string, org organisation, dbName string, db dashboard) error {
	// check api key exists
	if org.APIKey == "" {
		return errors.New("Missing API key for Org: " + strconv.Itoa(int(org.ID)))
	}
	// read in dashboard json
	b, err := ioutil.ReadFile(db.File)
	if err != nil {
		return err
	}
	bs := string(b)
	for k, v := range org.DashboardVars {
		bs = strings.Replace(bs, "#"+k+"#", v, -1)
	}
	data := "{\"overwrite\":true,\"dashboard\":" + bs + "}"
	// get existing if it exists
	log.Debugf("Posting dashboard \"%s\"", dbName)
	req, err := http.NewRequest("POST", ep+"api/dashboards/db/", strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+org.APIKey)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	log.Debugf("Posted as overwrite")
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	} else if resp.StatusCode == 404 {
		data = "{\"overwrite\":false,\"dashboard\":" + string(b) + "}"
		// get existing if it exists
		log.Debugf("Posting dashboard \"%s\"", dbName)
		req, err := http.NewRequest("POST", ep+"api/dashboards/db/", strings.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+org.APIKey)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return err
		}
		log.Debugf("Posted as new")
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return nil
		}
		return errors.New(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debugf("Failed response body: %s", string(body))
	log.Debugf("Failed request body: %s", string(data))
	return errors.New(resp.Status)
}
