/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addressobserver

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	//"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	clientv2alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/client/clientset/versioned/typed/broker/v2alpha1"
	"k8s.io/client-go/tools/clientcmd"
)

var log = logf.Log.WithName("addressobserver_v2alpha1activemqartemisaddress")

const AnnotationStatefulSet = "statefulsets.kubernetes.io/drainer-pod-owner"
const AnnotationDrainerPodTemplate = "statefulsets.kubernetes.io/drainer-pod-template"

type AddressObserver struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	statefulSetLister  appslisters.StatefulSetLister
	statefulSetsSynced cache.InformerSynced
	podLister          corelisters.PodLister
	podsSynced         cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func NewAddressObserver(
	kubeclientset kubernetes.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	namespace string) *AddressObserver {

	statefulSetInformer := kubeInformerFactory.Apps().V1().StatefulSets()
	podInformer := kubeInformerFactory.Core().V1().Pods()
	itemExponentialFailureRateLimiter := workqueue.NewItemExponentialFailureRateLimiter(5*time.Second, 300*time.Second)

	observer := &AddressObserver{
		kubeclientset:      kubeclientset,
		statefulSetLister:  statefulSetInformer.Lister(),
		statefulSetsSynced: statefulSetInformer.Informer().HasSynced,
		podLister:          podInformer.Lister(),
		podsSynced:         podInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(itemExponentialFailureRateLimiter, "StatefulSets"),
	}

	log.Info("Setting up event handlers")
	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs {
		AddFunc: observer.enqueueStatefulSet,
		UpdateFunc: func(old, new interface{}) {
			observer.enqueueStatefulSet(new)
		},
	})

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: observer.newPod,//only interested in new pods
	})

	return observer
}

func (c *AddressObserver) Run(stopCh <-chan struct{}) error {

    //if we don't need to block
    //then we probably don't need those defers
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting StatefulSet scaledown cleanup controller")

	// Wait for the caches to be synced before starting workers
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.statefulSetsSynced, c.podsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")
	go wait.Until(c.runWorker, time.Second, stopCh)

	log.Info("Started workers")
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

//is this necessary
func (c *AddressObserver) runWorker() {
}


func (c *AddressObserver) enqueueStatefulSet(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *AddressObserver) newPod(obj interface{}) {

    fmt.Printf("Observer got a new pod notif... %v", obj)
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		log.V(4).Info("Recovered deleted object " + object.GetName() + " from tombstone")
	}
	log.V(5).Info("Processing object: " + object.GetName())

    //AnnotationStatefulSet is from the drainer's controller.go, import it?
	stsNameFromAnnotation := object.GetAnnotations()[AnnotationStatefulSet]
	if stsNameFromAnnotation != "" {
		//ignore drainer pod
		return
	}

	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a StatefulSet, we should not do anything more
		// with it.
		if ownerRef.Kind != "StatefulSet" {
			return
		}

		sts, err := c.statefulSetLister.StatefulSets(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			log.V(4).Info("ignoring orphaned object " + object.GetSelfLink() + " of StatefulSet " + ownerRef.Name)
			return
		}

		if 0 == *sts.Spec.Replicas {
			log.V(5).Info("Name from ownerRef.Name not enqueueing Statefulset " + sts.Name + " as Spec.Replicas is 0.")
			return
		}

        //c.enqueueStatefulSet(sts)
		//now send the pod all crs
		c.checkCRsForPod(&object)
		return
	}
}

func (c *AddressObserver) checkCRsForPod(object *metav1.Object) {

	fmt.Printf("****Checking CRs for Pod: %v\n", object)

	apiLists, err := c.kubeclientset.Discovery().ServerResourcesForGroupVersion("broker.amq.io/v2alpha1")
//	apiLists, err := c.kubeclientset.Discovery().ServerResources()
	if err != nil {
		fmt.Printf("**** Error get server resources: %v\n", err)
		return;
	}
			for _, r := range apiLists.APIResources {
				fmt.Printf("**** checking up kind: %v\n", r.Kind)
				if r.Kind == "ActiveMQArtemisAddress" {
					fmt.Printf("**** found the resource!!! %v\n", r)
					fmt.Printf("string of it %v\n", r.String())
					fmt.Printf("categories: %s\n", r.Categories)
					fmt.Printf("resource name: %s\n", r.Name)
					fmt.Printf("resource size: %d\n", r.Size())
					fmt.Printf("verbs %v\n", r.Verbs)
					fmt.Printf("Group**** %v\n", r.Group)
					
				}
			}

//	result := v2alpha1.ActiveMQArtemisAddressList{}
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Error(err, "Error building kubeconfig: %s", err.Error())
	}

	brokerClient, err2 := clientv2alpha1.NewForConfig(cfg)
	if err2 != nil {
		log.Error(err2, "Error building brokerclient: %s", err2.Error())
	}

	addressInterface := brokerClient.ActiveMQArtemisAddresses((*object).GetNamespace())
	result, listerr := addressInterface.List(metav1.ListOptions{})
	if listerr != nil {
		fmt.Printf("**** Failed to get address resources %v\n", listerr)
		return		
	}
/*
	restClient := c.kubeclientset.CoreV1().RESTClient();
	err1 := restClient.Get().Namespace((*object).GetNamespace()).Resource("ActiveMQArtemisAddress").Do().Into(&result);
	
	if err1 != nil {
		fmt.Printf("**** Failed to get address resources %v\n", err1)
		return
	}
*/
	fmt.Printf("***** Found the result, %v\n", result.Items)

	for _, a := range result.Items {
		fmt.Printf("++++++++Address: %v, Queue: %v, RoutingType: %v\n", a.Spec.AddressName, a.Spec.QueueName, a.Spec.RoutingType)
	}
	
	fmt.Printf("******** done cr\n")
			
	/*
	for _, apiList := range apiLists {
        fmt.Printf("Checking up apiGroupVersion: %v\n", apiList.GroupVersion)
		if apiList.GroupVersion == "broker.amq.io/v2alpha1" {
			for _, r := range apiList.APIResources {
				fmt.Printf("**** checking up kind: %v\n", r.Kind)
				if r.Kind == "ActiveMQArtemisAddress" {
					fmt.Printf("**** found the resource!!! %v\n", r)
					fmt.Printf("string of it %v\n", r.String())
					fmt.Printf("categories: %s\n", r.Categories)
					fmt.Printf("resource name: %s\n", r.Name)
					fmt.Printf("resource size: %d\n", r.Size())
				}
			}
		}
	}
	*/
}
