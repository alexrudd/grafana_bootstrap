package main

import (
	"flag"
	"io/ioutil"
	"net/url"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	debug     = flag.Bool("debug-logging", false, "Debug loggin enabled")
	confFile  = flag.String("config", "grafana_bootstrap.yml", "The location of the bootstrap config to use")
	endpoint  = flag.String("endpoint", "http://localhost:3000/", "The Grafana api endpoint with credentials")
	userPass  = flag.String("user", "admin", "The Grafana user password")
	adminPass = flag.String("pass", "admin", "The Grafana admin password")
)

func init() {
	flag.Parse()

	if *debug == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

// MAIN
func main() {
	// Open and parse config file
	b, err := ioutil.ReadFile(*confFile)
	if err != nil {
		log.Errorln(err)
		return
	}
	c := config{}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		log.Errorln(err)
		return
	}

	// endpoints
	adminEndpoint, err := url.Parse(*endpoint)
	if err != nil {
		log.Errorln(err)
		return
	}
	adminEndpoint.User = url.UserPassword(*userPass, *adminPass)

	// Check organisations exists
	for _, org := range c.Organisations {
		o, err := doesOrgExist(adminEndpoint.String(), org.Name)
		if err != nil {
			o, err = createOrg(adminEndpoint.String(), org.Name)
			if err != nil {
				log.Errorln(err)
				return
			}
			log.Debugf("Created organisation \"%s\"", org.Name)
		}
		log.Debugf("Organisation \"%s\" exists with ID %v", org.Name, o.ID)
		org.ID = o.ID

		// Datasources
		for _, orgDs := range org.Datasources {
			log.Debugf("Creating/Updating \"%s\" for Org %v", orgDs, org.ID)
			err = createUpdateDatasource(*endpoint, org, orgDs, c.Datasources[orgDs])
			if err != nil {
				log.Errorln(err)
				return
			}
		}
		// Dashboards
		for _, orgDb := range org.Dashboards {
			log.Debugf("Creating/Updating db \"%s\" for Org %v", orgDb, org.ID)
			err = createUpdateDashboard(*endpoint, org, orgDb, c.Dashboards[orgDb])
			if err != nil {
				log.Errorln(err)
			}
		}
	}
}
