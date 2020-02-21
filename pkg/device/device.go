package device

import (
	"sync"
	"strconv"
	_"strings"	
	"errors"

	"k8s.io/klog"
	"github.com/jwzl/edgeOn/common"
	"github.com/jwzl/edgedev/pkg/config"
	"github.com/jwzl/edgedev/pkg/transfer"
)

const (
	DEVICE_STATE_ONLINE	= "online"
	DEVICE_STATE_OFFLINE	= "offline"
)
var (
	gDeviceTwin *common.DeviceTwin
)

type Device struct {
	//notify channel for device property update.
	NotifyCh 			chan []string		
	/* Device is online/offline? */
	State 				string
	/* transfer handle. etc. mqtt.*/	
	transferHandle		transfer.Transfer

	deviceMutex			*sync.Mutex
	DeviceTwin 			*common.DeviceTwin
}
func initDevice(conf *config.DeviceConfig) (*Device, error) {
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

	//create device data struct.
	var devMutex	sync.Mutex
	nchn := make(chan []string, 120)
	dev :=	&Device{
		NotifyCh: nchn,
		State: DEVICE_STATE_OFFLINE,
		deviceMutex: &devMutex,
		DeviceTwin: deviceTwin,
	}

	dev.transferHandle = transfer.NewClient(brokerUrl, dev.onMessageArrived)
	//init transfer client and start to connect the server.
	dev.transferHandle.InitAndConnect()
	return dev, nil
}

/*
* On message arrived.
*/
func (dev *Device) onMessageArrived(topic string payload []byte){

}
/*
* Match for incoming request.
*/
func (dev *Device) Match(msgTwin *common.DeviceTwin) bool {

	if msgTwin.ID != dev.DeviceTwin.ID {
		return false
	}

	dev.State = DEVICE_STATE_ONLINE
	return true
}


/*
* Update properties of this device.
*/
func (dev *Device) UpdateProps(msgTwin *common.DeviceTwin) error {
	propArray []string

	if msgTwin.ID != dev.DeviceTwin.ID {
		klog.Warningf("unexpected deviceID %d", msgTwin.ID)
		return errors.New("device is not matched.")
	}

	if dev.State != DEVICE_STATE_ONLINE {
		return errors.New("device is offline")
	}

	savedDesired := dev.DeviceTwin.Properties.Desired
	newDesired := msgTwin.Properties.Desired
	propArray = make([]string, 0)
	dev.deviceMutex.Lock()
	/*
	* update all value.
	*/
	for index, _ := range newDesired {
		for key, _ := range savedDesired {
			if newDesired[index].Name == savedDesired[key].Name {
				savedDesired[key].Value = newDesired[index].Value 	
				propArray = append(propArray, newDesired[index].Name)
			}
		}
	}
	dev.deviceMutex.Unlock()
	
	//notify the channel.
	dev.NotifyCh <- propArray

	return nil                            
}

/*
* Sync device properties to edge.
*/
func (dev *Device) SyncDeviceProperties(properties map[string][]byte) error {

	syncProps := make([]common.TwinProperty, 0)
	//sync to reported properties.
	dev.deviceMutex.Lock()
	savedReported := dev.DeviceTwin.Properties.Reported
	for name, _ := range properties {
		for key, _ := range savedReported {
			if savedReported[key].Name == name {
				savedReported[key].Value = properties[name]
				syncProps = append(syncProps, savedReported[key])
			}
		}
	}
	dev.deviceMutex.Unlock()

	//build device message.
	deviceMessage := dev.buildDeviceReportMessage(syncProps)

	//send to mqtt module to send this message.
	return nil
}

func (dev *Device) GetDeviceID() string {
	return dev.DeviceTwin.ID
}

/*
* Get property value.
*/
func (dev *Device) GetPropertyDesiredValue(name string) ([]byte, error) {
	savedDesired := dev.DeviceTwin.Properties.Desired
	for _, prop := range savedDesired {
		if prop.Name == name {
			return prop.Value, nil
		}
	}
	
	return nil, errors.New("not found")
}

/*
* build device report message.
*/
func (dev *Device) buildDeviceReportMessage(reported []common.TwinProperty) *common.DeviceMessage {
	
	deviceTwin := common.DeviceTwin{
		ID: dev.DeviceTwin.ID,
		Name: dev.DeviceTwin.Name,
		Properties:	common.DeviceTwinProperties{
			Reported: reported,
		},
	}

	return &common.DeviceMessage{
		Twin: deviceTwin,
	}
}
