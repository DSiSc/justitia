For now, we use filebeat, metricbeat, elasticsearch and kibana to facilitate log analysis and metrics monitoring.

There are four components:

- filebeat: to collect logs from log file, and send them to elasticsearch;
- metricbeat: to collect metrics from certain api, and send them to elasticsearch;
- elasticsearch: receive and store logs and metrics, and provide query api;
- kibana: to show filted logs and visualized metrics.

So, filebeat and metricbeat usually be deployed on the same host with justitia nodes; 
Whereas elasticsearch and kibana usually be deployed on seperate host or cluster.

## Deploying filebeat and metricbeat

`docker-compose.yml` and configuration files are under `agents`.

1. Before start them, configure them properly.
  * Set proper elasticsearch host in `filebeat.yml` and `metricbeat.yml`.
  * Set correct log file path of justitia in `justitia.yml`.
  * Set correct ip of justitia in `prometheus.yml`.
2. Use `docker-compose up -d` to start filebeat and metricbeat.

## Deploying elasticsearch and kibana

1. You can find `docker-compose.yml` under `center`.
  * data files of es and kibana are exposed to host through mounted volumes, config it as you wish.
2. Use `docker-compose up -d` to start elasticsearch and kibana.
3. Open `http://kibana:5601` in browser and start to create indexs, discovers, dashboards, or just simply import `kibana-import.json`.
