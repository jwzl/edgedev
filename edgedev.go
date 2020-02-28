package main

import (

	"k8s.io/klog"
	"github.com/jwzl/edgedev/pkg/config"
	"github.com/jwzl/edgedev/pkg/device"
)



func main() {

	deviceProfile, err := config.GetDeviceProfileFile("./conf/deviceProfile.json")
	if err != nil {
		klog.Warningf("get device profile error with err %v", err)
		return
	}	
	klog.Infof("device profile is", deviceProfile)

	dev, err := device.InitDevice(deviceProfile)
	if err != nil {
		klog.Warningf("init device with err %v", err)
		return
	}	

	id := dev.GetDeviceID()
	klog.Infof("device id is", id )
	
	dev.WaitDeviceOnline()
	klog.Infof("device id is online")
	
	for {
		props, ok := <- dev.GetUpdateCh()
		if !ok {
			klog.Warningf("channel is closed")
			return
		}

		for _ , name := range props {
			varray, err := dev.GetPropertyDesiredValue(name)
			if err != nil {
				klog.Warningf("%s is not found, ignored", name)
			}
			switch name {	
			case "led_pin0":			
				klog.Infof("led_pin0 property update to ", varray)
				//value:= varray[0]
				
				reported := map[string][]byte{name:varray}

				klog.Infof("Send the reported property ")
				dev.SyncDeviceProperties(reported)
			default:
				klog.Infof("No such property ")
			}
		} 
	}
}
