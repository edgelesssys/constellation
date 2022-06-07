package cloudtypes

import "errors"

// Instance is a gcp instance.
type Instance struct {
	PublicIP  string
	PrivateIP string
}

// Instances is a map of gcp Instances. The ID of an instance is used as key.
type Instances map[string]Instance

// IDs returns the IDs of all instances of the Constellation.
func (i Instances) IDs() []string {
	var ids []string
	for id := range i {
		ids = append(ids, id)
	}
	return ids
}

// PublicIPs returns the public IPs of all the instances of the Constellation.
func (i Instances) PublicIPs() []string {
	var ips []string
	for _, instance := range i {
		ips = append(ips, instance.PublicIP)
	}
	return ips
}

// PrivateIPs returns the private IPs of all the instances of the Constellation.
func (i Instances) PrivateIPs() []string {
	var ips []string
	for _, instance := range i {
		ips = append(ips, instance.PrivateIP)
	}
	return ips
}

// GetOne return anyone instance out of the instances and its ID.
func (i Instances) GetOne() (string, Instance, error) {
	for id, instance := range i {
		return id, instance, nil
	}
	return "", Instance{}, errors.New("map is empty")
}

// GetOthers returns all instances but the one with the handed ID.
func (i Instances) GetOthers(id string) Instances {
	others := make(Instances)
	for key, instance := range i {
		if key != id {
			others[key] = instance
		}
	}
	return others
}
