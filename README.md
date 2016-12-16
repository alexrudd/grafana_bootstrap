# grafana_bootstrap
A Grafana configuration tool for creating organisations, datasources, and dashboards

Use a YAML configuration file to add and edit Garfana organisations, datasources, and dashboards. Use org specific variables to customise generic dashboards for each organisation.

## Example

Create two organisations which share three datasources and three dashboards:

```yml
organisations:
  # Internal Devs - needs full visability and detailed monitoring
  # from all environments and all application logs.
  - name: "Internal Devs"
    apiKey: ""
    datasources:
      - production_monitoring
      - development_monitoring
      - application_logs
    dashboards:
      - detailed_environment_monitoring
      - logs_dash
      - environment_availability
    dashboardVars:
      logFilterRegex: ".*"

  # Internal Business - Only cares about production env availability
  # and logs from "frontend_app".
  - name: "Internal Business"
    apiKey: ""
    datasources:
      - production_monitoring
      - application_logs
    dashboards:
      - logs_dash
      - environment_availability
    dashboardVars:
      logFilterRegex: "^frontend_app$"

datasources:

  production_monitoring:
    type: "prometheus"
    access: "proxy"
    url: "http://prod.monitoring.example.com:9090"
    isDefault: true

  development_monitoring:
    type: "prometheus"
    access: "proxy"
    url: "http://dev.monitoring.example.com:9090"
    isDefault: false

  application_logs:
    type: "elasticsearch"
    access: "proxy"
    url: "http://es.monitoring.example.com:9200"
    database: "[logs-app-*]YYYY-MM-DD"
    isDefault: false
    jsonData:
      esVersion: 2
      interval: "Daily"
      timeField: "@timestamp"

dashboards:
  # Dashboards will have any/all instances of the string
  # '#logFilterRegex#' replaced with the value defined in their
  # organisation's 'dashboardVars' section.

  detailed_environment_monitoring:
    file: ./dashboards/detailed_environment_monitoring.json

  logs_dash:
    file: ./dashboards/logs_dash.json

  environment_availability:
    file: ./dashboards/environment_availability.json
```

Run Grafana bootstrap:

```bash
./grafana_bootstrap \
  -endpoint="https://grafana.monitoring.example.com" \
  -config="example_config.yml" \
  -pass="adminpass" \
  -debug-logging \
```

**NOTE:** When adding new Organisations, the first run of grafana_bootstrap will fail. This is because to add datasources and dashboards, the bootstrapper needs to know and API key for the Organisation, but a brand new organisation doesn't have an API key yet!

API keys can only be created by an admin user, through the Grafana web ui.

...Include the now created API keys into the 'organisation' config, like so:

```yml
organisations:
  - name: "Internal Devs"
    apiKey: "eyJrIjoiczFHT2I3clREcmVZNzBDdUdnMlRxOFd6VWdITmpMRjEiLCJuIjoibWFuYWdlbWVudCIsImlkIjoxfQ=="
```

Running the bootstrapper again should now succeed.

## Dashboard Variables

I found it useful to be able to customise generic dashboard templates by inserting organisation specific variables. The can be done by mapping keys to values in the 'dashboardVars' section of the organisation config.

Each json dashboard template will have a global find and replace performed on it, for the variable key surrounded by '#'s (`#myVarKey#`) and the corresponding value.

## Issues

Besides the hassle with having to manually create each organisations API key, grafana_bootstrap stores no state. This means that dashboard, datasource, or organisation name changes will not be picked up; and grafana_bootstrap will assume these are new resources which need to be created.

This is mostly a very dumb and hacky tool, feel free to contribute.

## Building

Built and tested using `go version go1.7.4 linux/amd64`

```bash
go build
```
