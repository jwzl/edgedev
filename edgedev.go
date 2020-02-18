package main

import (

	"k8s.io/klog"
	"github.com/jwzl/edgedev/pkg/config"
)



func main() {

	deviceProfile, err := config.GetDeviceProfileFile("./conf/deviceProfile.json")
	if err != nil {
		klog.Warningf("get device profile error with err %v", err)
		return
	}	

	klog.Infof("device profile is", deviceProfile)
}
