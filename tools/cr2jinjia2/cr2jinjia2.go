package main

import (
    "flag"
    "fmt"
	"io/ioutil"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v3alpha1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/cr2jinja2"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"bytes"
)

/*
Usage cr2jinja2 [options] broker_cr.yaml
options:
-e the environment variable
-o the output file
*/
func main() {

	envVar := flag.String("e", "", "env var")
	outputVar := flag.String("o", "default.yaml", "output file")

    flag.Parse()

    fmt.Println("-e:", *envVar)
    fmt.Println("-o", *outputVar)
    fmt.Println("tail: ", flag.Args())
//    fmt.Println("main", os.Args[0])
	if len(flag.Args()) != 1 {
		fmt.Println("you need to pass in the cr file")
		return
	}
	crFile := flag.Args()[0];
	fmt.Println("cr file " + crFile)
	yamlFile, err := ioutil.ReadFile(crFile)
	
	if (err != nil) {
		fmt.Println("err", err)
		return
	}

	brokerCr := v3alpha1.ActiveMQArtemis{}
	
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlFile), 1024)
	err = decoder.Decode(&brokerCr)

	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
		return
	}

	fmt.Printf("apiVersion: %v\n", brokerCr.APIVersion)
	fmt.Printf("plan: %v\n", brokerCr.Spec.DeploymentPlan.Image)
	fmt.Printf("addressting; %v\n", brokerCr.Spec.AddressSettings.AddressSetting[0].DeadLetterAddress)
	
	cr2jinja2.MakeBrokerCfgOverrides(&brokerCr, envVar, outputVar);

	fmt.Println("============================================")

}