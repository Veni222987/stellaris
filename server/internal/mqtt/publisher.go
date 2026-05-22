// Package mqtt 封装调度中心与 Planet 节点之间通过 MQTT 通信的两端。
package mqtt

import (
	"encoding/json"
	"errors"

	mqttc "github.com/eclipse/paho.mqtt.golang"
	"github.com/stellaris/stellaris/shared/protocol"
)

type Publisher struct{ client mqttc.Client }

func NewPublisher(client mqttc.Client) *Publisher { return &Publisher{client: client} }

// PublishTask 把任务下发到指定 Planet 的 task topic。QoS=1 保证至少送达一次。
func (p *Publisher) PublishTask(gid, planetUUID string, msg protocol.TaskMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if !p.client.IsConnected() {
		return errors.New("mqtt broker disconnected")
	}
	t := p.client.Publish(protocol.TaskTopic(gid, planetUUID), 1, false, payload)
	t.Wait()
	return t.Error()
}
