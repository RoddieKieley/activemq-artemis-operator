package cr2jinja2

import (
	brokerv3alpha1 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v3alpha1"
)

/* return a yaml string */
func MakeBrokerCfgOverrides(customeResource *brokerv3alpha1.ActiveMQArtemis, envVar *string, output *string) string {
	return "broker_home: /opt/artemis-2.14.0"
}