package secrets

import (
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/random"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/selectors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("package secrets")
var CredentialsNameBuilder namer.NamerData
var ConsoleNameBuilder namer.NamerData
var NettyNameBuilder namer.NamerData

func MakeStringDataMap(keyName string, valueName string, key string, value string) map[string]string {

	if 0 == len(key) {
		key = random.GenerateRandomString(8)
	}

	if 0 == len(value) {
		value = random.GenerateRandomString(8)
	}

	stringDataMap := map[string]string{
		keyName:   key,
		valueName: value,
	}

	return stringDataMap
}

//func MakeSecret(customResource *brokerv2alpha1.ActiveMQArtemis, secretName string, stringData map[string]string) corev1.Secret {
func MakeSecret(namespacedName types.NamespacedName, secretName string, stringData map[string]string) corev1.Secret {

	secretDefinition := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    selectors.LabelBuilder.Labels(),
			Name:      secretName,
			Namespace: namespacedName.Namespace,
		},
		StringData: stringData,
	}

	return secretDefinition
}

//func NewSecret(customResource *brokerv2alpha1.ActiveMQArtemis, secretName string, stringData map[string]string) *corev1.Secret {
func NewSecret(namespacedName types.NamespacedName, secretName string, stringData map[string]string) *corev1.Secret {

	secretDefinition := MakeSecret(namespacedName, secretName, stringData)

	return &secretDefinition
}

func Create(owner metav1.Object, namespacedName types.NamespacedName, stringDataMap map[string]string, client client.Client, scheme *runtime.Scheme) *corev1.Secret {

	var err error = nil
	secretDefinition := NewSecret(namespacedName, namespacedName.Name, stringDataMap)

	if err = resources.Retrieve(namespacedName, client, secretDefinition); err != nil {
		if errors.IsNotFound(err) {
			err = resources.Create(owner, namespacedName, client, scheme, secretDefinition)
		}
	}

	return secretDefinition
}

//func RetrieveValues(namespacedName types.NamespacedName, secretDefinition *corev1.Secret, client client.Client, scheme *runtime.Scheme) (error, stringDataMap map[string]string)  {
//
//	var err error = nil
//
//	namespacedName := types.NamespacedName{
//		Name:      secretName,
//		Namespace: currentStatefulSet.Namespace,
//	}
//	// Attempt to retrieve the secret
//	stringDataMap := make(map[string]string)
//	for k := range *envVars {
//		stringDataMap[k] = (*envVars)[k]
//	}
//	secretDefinition := secrets.NewSecret(namespacedName, secretName, stringDataMap)
//	if err = resources.Retrieve(namespacedName, client, secretDefinition); err != nil {
//		if errors.IsNotFound(err) {
//			log.Info("sourceEnvVarFromSecret did not find secret " + secretName)
//			requestedResources = append(requestedResources, secretDefinition)
//		}
//	} else { // err == nil so it already exists
//		// Exists now
//		// Check the contents against what we just got above
//		log.Info("sourceEnvVarFromSecret did found secret " + secretName)
//
//		var needUpdate bool = false
//		for k := range *envVars {
//			elem, ok := secretDefinition.Data[k]
//			if 0 != strings.Compare(string(elem), (*envVars)[k]) || !ok {
//				log.Info("Secret exists but not equals, or not ok", "ok?", ok)
//				secretDefinition.Data[k] = []byte((*envVars)[k])
//				needUpdate = true
//			}
//		}
//
//	return err
//}