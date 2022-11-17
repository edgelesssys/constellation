.PHONY: containers

containers:
	docker build ./filebeat -t ghcr.io/edgelesssys/constellation/filebeat-debug
	docker build ./logstash -t ghcr.io/edgelesssys/constellation/logstash-debug
	docker push ghcr.io/edgelesssys/constellation/filebeat-debug
	docker push ghcr.io/edgelesssys/constellation/logstash-debug
