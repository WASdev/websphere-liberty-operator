/*
  Copyright contributors to the WASdev project.

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

package controllers

import (
	"context"

	webspherelibertyv1 "github.com/WASdev/websphere-liberty-operator/api/v1"

	appstacksutils "github.com/application-stacks/runtime-component-operator/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ handler.EventHandler = &EnqueueRequestsForCustomIndexField{}

const (
	indexFieldImageStreamName = "spec.applicationImage"
)

// EnqueueRequestsForCustomIndexField enqueues reconcile Requests for WebSphereLiberty Applications if the app is relying on
// the modified resource
type EnqueueRequestsForCustomIndexField struct {
	handler.Funcs
	Matcher CustomMatcher
}

// Create implements EventHandler
func (e *EnqueueRequestsForCustomIndexField) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Object, evt.Object, q)
}

// Update implements EventHandler
func (e *EnqueueRequestsForCustomIndexField) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.ObjectNew, evt.ObjectNew, q)
}

// Delete implements EventHandler
func (e *EnqueueRequestsForCustomIndexField) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Object, evt.Object, q)
}

// Generic implements EventHandler
func (e *EnqueueRequestsForCustomIndexField) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.handle(evt.Object, evt.Object, q)
}

// handle common implementation to enqueue reconcile Requests for applications
func (e *EnqueueRequestsForCustomIndexField) handle(evtMeta metav1.Object, evtObj runtime.Object, q workqueue.RateLimitingInterface) {
	apps, _ := e.Matcher.Match(evtMeta)
	for _, app := range apps {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: app.Namespace,
				Name:      app.Name,
			}})
	}
}

// CustomMatcher is an interface for matching apps that satisfy a custom logic
type CustomMatcher interface {
	Match(metav1.Object) ([]webspherelibertyv1.WebSphereLibertyApplication, error)
}

// ImageStreamMatcher implements CustomMatcher for Image Streams
type ImageStreamMatcher struct {
	Klient          client.Client
	WatchNamespaces []string
}

// Match returns all applications using the input ImageStreamTag
func (i *ImageStreamMatcher) Match(imageStreamTag metav1.Object) ([]webspherelibertyv1.WebSphereLibertyApplication, error) {
	apps := []webspherelibertyv1.WebSphereLibertyApplication{}
	var namespaces []string
	if appstacksutils.IsClusterWide(i.WatchNamespaces) {
		nsList := &corev1.NamespaceList{}
		if err := i.Klient.List(context.Background(), nsList, client.InNamespace("")); err != nil {
			return nil, err
		}
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	} else {
		namespaces = i.WatchNamespaces
	}
	for _, ns := range namespaces {
		appList := &webspherelibertyv1.WebSphereLibertyApplicationList{}
		err := i.Klient.List(context.Background(),
			appList,
			client.InNamespace(ns),
			client.MatchingFields{indexFieldImageStreamName: imageStreamTag.GetNamespace() + "/" + imageStreamTag.GetName()})
		if err != nil {
			return nil, err
		}
		apps = append(apps, appList.Items...)
	}

	return apps, nil
}
