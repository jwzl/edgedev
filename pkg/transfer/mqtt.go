package transfer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"
	"crypto/tls"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	// TokenWaitTime to wait
	TokenWaitTime = 120 * time.Second

	SubTopics = []string{
		//"$hw/events/upload/#",
		"$hw/events/device/#",
	}
)

type MessageArrivedFunc func(topic string, payload []byte)
// Client struct
type Client struct {
	MQTTUrl string
	PubCli  MQTT.Client
	SubCli  MQTT.Client
	onSubMessageFunc	MessageArrivedFunc
}


func NewMqttClient(url string, subFunc MessageArrivedFunc) *Client {

	return &Client{
		MQTTUrl: url,
		onSubMessageFunc: subFunc,
	}
}

// CheckClientToken checks token is right
func CheckClientToken(token MQTT.Token) (bool, error) {
	if token.Wait() && token.Error() != nil {
		return false, token.Error()
	}
	return true, nil
}

// LoopConnect connect to mqtt server
func (mq *Client) LoopConnect(clientID string, client MQTT.Client) {
	for {
		klog.Infof("start connect to mqtt server with client id: %s", clientID)
		token := client.Connect()
		klog.Infof("client %s isconnected: %v", clientID, client.IsConnected())
		if rs, err := CheckClientToken(token); !rs {
			klog.Errorf("connect error: %v", err)
		} else {
			return
		}
		time.Sleep(5 * time.Second)
	}
}


func (mq *Client) onPubConnectionLost(client MQTT.Client, err error) {
	klog.Errorf("onPubConnectionLost with error: %v", err)
	go mq.InitPubClient()
}

func (mq *Client) onSubConnectionLost(client MQTT.Client, err error) {
	klog.Errorf("onSubConnectionLost with error: %v", err)
	go mq.InitSubClient()
}

func (mq *Client) onSubConnect(client MQTT.Client) {
	for _, t := range SubTopics {
		token := client.Subscribe(t, 1, mq.OnSubMessageReceived)
		if rs, err := CheckClientToken(token); !rs {
			klog.Errorf("edge-hub-cli subscribe topic: %s, %v", t, err)
			return
		}
		klog.Infof("device-hub-cli subscribe topic to %s", t)
	}
}

// OnSubMessageReceived msg received callback
func (mq *Client) OnSubMessageReceived(client MQTT.Client, message MQTT.Message) {
	// for "$hw/events/device/#", send to twin
	
	if strings.HasPrefix(message.Topic(), "$hw/events/device") {
		if mq.onSubMessageFunc != nil {
			klog.Info("message arrived! ")
			mq.onSubMessageFunc(message.Topic(), message.Payload())
		}
	}  
}

//Send Publish message to edge.
func (mq *Client) Send(topic string, payload []byte) error {
	token := mq.PubCli.Publish(topic, 1, false, payload)
	if token.WaitTimeout(TokenWaitTime) && token.Error() != nil {
		return token.Error()
	} 

	return nil
}

// HubClientInit create mqtt client config
func (mq *Client) HubClientInit(server, clientID, username, password string) *MQTT.ClientOptions {
	opts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientID).SetCleanSession(true)
	if username != "" {
		opts.SetUsername(username)
		if password != "" {
			opts.SetPassword(password)
		}
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)
	return opts
}

// InitSubClient init sub client
func (mq *Client) InitSubClient() {
	timeStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	right := len(timeStr)
	if right > 10 {
		right = 10
	}

	subID := fmt.Sprintf("client-sub-%s", timeStr[0:right])
	subOpts := mq.HubClientInit(mq.MQTTUrl, subID, "", "")
	subOpts.OnConnect = mq.onSubConnect
	subOpts.AutoReconnect = false
	subOpts.OnConnectionLost = mq.onSubConnectionLost
	mq.SubCli = MQTT.NewClient(subOpts)
	mq.LoopConnect(subID, mq.SubCli)
	klog.Info("finish client sub")
}


// InitPubClient init pub client
func (mq *Client) InitPubClient() {
	timeStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	right := len(timeStr)
	if right > 10 {
		right = 10
	} 

	pubID := fmt.Sprintf("client-pub-%s", timeStr[0:right])
	pubOpts := mq.HubClientInit(mq.MQTTUrl, pubID, "", "")
	pubOpts.OnConnectionLost = mq.onPubConnectionLost
	pubOpts.AutoReconnect = false
	mq.PubCli = MQTT.NewClient(pubOpts)
	mq.LoopConnect(pubID, mq.PubCli)
	klog.Info("finish client pub")
}

// init and connect
func (mq *Client) InitAndConnect(){
	mq.InitSubClient()
	mq.InitPubClient()
}
