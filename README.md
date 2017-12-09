# UnavailabilityCounter

~~~ shell
$ docker run -d --restart=always --name <name> -p <port>:9143 -v <dataDirectory>:/data -v <confDirectory>:/go/src/app shakapark/unavailabilitycounter:1.08
~~~

## Prometheus Configuration

~~~ shell
scrape_configs:
  - job_name: indispo
    metrics_path: /probe
    static_configs:
      - targets: ['<ip>:<port>']
~~~

## On Grafana

To view the state of probes, use the prometheus metrics :
~~~ shell
probe_success_<InstanceName>_<GroupeName>{target="<targetURL|targetIP>"} #with prometheus datasource
~~~

Add this program to Grafana as a Prometheus datasource. And enter this request to get the percent of unavailability :
~~~ shell
time{instance=<InstanceName>}                                            #with new datasource 
~~~

If you have several groups in your configuration, you can separate unavailability :
~~~ shell
time{instance=<InstanceName>,group=<GroupName>}                          #with new datasource 
~~~
