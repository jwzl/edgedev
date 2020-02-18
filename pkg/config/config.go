package config

import(
	"io/ioutil"
	"encoding/json"

	"k8s.io/klog"
)

type DeviceConfig struct{
	// device id	
	DeviceID		string			`json:"id"`
	//device name
	Name	string 					`json:"name,omitempty"`
	// device description
	Description		string			`json:"description,omitempty"`
	// device metadata  
	MetaData	map[string]string	`json:"metadata,omitempty"`
	//all properties
	Properties	[]DeviceProperty			`json:"properties,omitempty"`
}

type DeviceProperty struct {
	Name	string 					`json:"name,omitempty"`
	Type	string 					`json:"type,omitempty"`
	/* property meta data.*/
	MetaData	map[string]string	`json:"metadata,omitempty"`
}


func GetDeviceProfileFile(filepath string)(*DeviceConfig, error){
	deviceProfile := &DeviceConfig{}

	jsonData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(jsonData, deviceProfile)
	if err != nil {
		return nil, err
	}

	return deviceProfile, nil
}
