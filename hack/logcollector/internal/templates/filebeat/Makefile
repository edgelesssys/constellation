TIMEOUT := 1200s
RELEASE := filebeat
CHART := elastic/filebeat

add:
	helm repo add elastic https://helm.elastic.co
	helm repo update

install:
	helm install $(RELEASE) $(CHART) \
	--wait --timeout=$(TIMEOUT) --values values.yml
	
remove:
	helm uninstall $(RELEASE)
