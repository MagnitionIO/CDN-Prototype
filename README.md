# CDN-Prototype

## How to run

```bash
git clone https://github.com/yazhuo/CDN-Prototype.git
```

In your working directory,
```bash
docker-compose up --build
```

## Metrics
Raw metrics: `http://127.0.0.1:9090/client/metrics`

## Logs
The varnish misses and hits are recorded in `logs/cdn_client.log`

## Configurations
`docker-compose.yml`: The default topology is one client, two Varnish caches, one ATS cache, and one origin

`docker-compose-varnishobly.yml`: Disable ATS, the topology is one client, two Varnish caches, and one origin.

You can update the workload for client service in the `docker-compose.yml`:
```
command:
      - "--server-addr=:9090" # server address for prometheus
      - "--origin-addrs=frontend1:80,frontend2:80" # server address for varnish
      - "--wiki-file=/data/wiki_1k.csv"
      - "--log-level=debug"
      - "--cpus=1"
```