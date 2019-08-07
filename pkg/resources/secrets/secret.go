package secrets

import (
	brokerv2alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/random"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("package secrets")

func MakeUserPasswordStringData(keyName string, valueName string, key string, value string) map[string]string {

	if 0 == len(key) {
		key = random.GenerateRandomString(8)
	}

	if 0 == len(value) {
		value = random.GenerateRandomString(8)
	}

	stringDataMap := map[string]string {
		keyName: key,
		valueName: value,
	}

	return stringDataMap
}

func MakeUserPasswordSecret(customResource *brokerv2alpha1.ActiveMQArtemis, secretName string, stringData map[string]string) corev1.Secret {

	userPasswordSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    selectors.LabelBuilder.Labels(),
			Name:      secretName,
			Namespace: customResource.Namespace,

		},
		StringData: stringData,
	}

	return userPasswordSecret
}

func NewUserPasswordSecret(customResource *brokerv2alpha1.ActiveMQArtemis, secretName string, stringData map[string]string) *corev1.Secret {

	userPasswordSecret := MakeUserPasswordSecret(customResource, secretName, stringData)

	return &userPasswordSecret
}
