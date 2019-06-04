/*
Copyright 2019 Google LLC

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

package pubsubsource

import (
	"context"
	"github.com/GoogleCloudPlatform/cloud-run-events/pkg/apis/events/v1alpha1"
	eventsinformers "github.com/GoogleCloudPlatform/cloud-run-events/pkg/client/informers/externalversions/events/v1alpha1"
	"github.com/GoogleCloudPlatform/cloud-run-events/pkg/pubsubutil"
	"github.com/GoogleCloudPlatform/cloud-run-events/pkg/reconciler"
	"github.com/knative/pkg/controller"
	"github.com/knative/pkg/tracker"
	appsv1informers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "cloud-run-events-pubsub-source-controller"
)

// NewController initializes the controller and is called by the generated code
// Registers eventhandlers to enqueue events
func NewController(
	opt reconciler.Options,
	deploymentInformer appsv1informers.DeploymentInformer,
	sourceInformer eventsinformers.PubSubSourceInformer,
	raPubSubSourceImage string,
	googleCreds string,
) *controller.Impl {

	c := &Reconciler{
		Base:                reconciler.NewBase(opt, controllerAgentName),
		deploymentLister:    deploymentInformer.Lister(),
		sourceLister:        sourceInformer.Lister(),
		pubSubClientCreator: pubsubutil.GcpPubSubClientCreatorWithCreds(context.Background(), googleCreds),
		receiveAdapterImage: raPubSubSourceImage,
	}
	impl := controller.NewImpl(c, c.Logger, ReconcilerName)

	c.Logger.Info("Setting up event handlers")
	sourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("PubSubSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	c.tracker = tracker.New(impl.EnqueueKey, opt.GetTrackerLease()) // TODO: use tracker.

	return impl
}