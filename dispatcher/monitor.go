package dispatcher

import "encoding/json"

type monitor struct {
	*externalSubscriber
}

func NewMonitor(address string) *monitor {
	return &monitor{NewExternalSubscriber(address)}
}

func (m *monitor) marshal(n Notification) (string, error) {
	jsonNotification, err := json.Marshal(n)

	if err != nil {
		return "", err
	}

	return string(jsonNotification), err
}

func (m *monitor) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSubscriberPayload(m))
}
