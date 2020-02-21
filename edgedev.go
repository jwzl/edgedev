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

	device, err := device.InitDevice(deviceProfile)
	if err != nil {
		klog.Warningf("init device with err %v", err)
		return
	}	

	id := device.GetDeviceID()
	klog.Infof("device id is", id )
	
}
