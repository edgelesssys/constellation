TIMEOUT := 1200s
RELEASE := logstash
CHART := elastic/logstash

add:
	helm repo add elastic https://helm.elastic.co
	helm repo update

install:
	helm install $(RELEASE) $(CHART) \
	--wait --timeout=$(TIMEOUT) --values values.yaml \
	--version 8.5.1

remove:
	helm uninstall $(RELEASE)
