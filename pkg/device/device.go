package device

import (
	"strconv"
	_"strings"	
	"errors"

	_"k8s.io/klog"
	"github.com/jwzl/edgeOn/common"
	"github.com/jwzl/edgedev/pkg/config"
	
)

var (
	gDeviceTwin *common.DeviceTwin
)

func initDevice(conf *config.DeviceConfig) (*common.DeviceTwin, error) {
	var length int

	if conf == nil {
		return nil, errors.New("conf is empty.")
	}

	deviceTwin := &common.DeviceTwin{
		ID: 	conf.DeviceID,
		Name:	conf.Name,
		Description: conf.Description,	
	}

	deviceTwin.MetaData = make([]common.MetaType, 0)
	for name, value := range conf.MetaData {
		meta := common.MetaType{
			Name: name,
			Value: value,
		} 
		deviceTwin.MetaData = append(deviceTwin.MetaData, meta)
	}
	
	deviceTwin.Properties.Desired = make([]common.TwinProperty, 0)
	deviceTwin.Properties.Reported = make([]common.TwinProperty, 0)
	
	//update all properties.
	for _, prop := range conf.Properties {
		deviceProp := common.TwinProperty{
			Name: prop.Name,
			Type: prop.Type,
			MetaData: make([]common.MetaType, 0),
		}
	
		switch deviceProp.Type {
		case common.TWIN_PROP_VALUE_TYPE_CHAR:
			length = 1
		case common.TWIN_PROP_VALUE_TYPE_UINT8:
			length = 1
		case common.TWIN_PROP_VALUE_TYPE_UINT16:
			length = 2	
		case common.TWIN_PROP_VALUE_TYPE_UINT32:
			length = 4
		case common.TWIN_PROP_VALUE_TYPE_UINT64:
			length = 8
		default:
			length = 4
		}

		//add the metadata.
		for name, value := range prop.MetaData {
			meta := common.MetaType{
				Name: name,
				Value: value,
			} 
			if name == "length" {
				length, _ = strconv.Atoi(value)
			}
			deviceProp.MetaData = append(deviceProp.MetaData, meta)
		}
		
		deviceProp.Value = make([]byte, length)
		deviceTwin.Properties.Reported = append(deviceTwin.Properties.Reported, deviceProp)
		deviceTwin.Properties.Desired = append(deviceTwin.Properties.Desired, deviceProp)
	}

	return deviceTwin, nil
}
