/*
Copyright 2020 CaiCloud, Inc. All rights reserved.

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

package filters

import (
	"k8s.io/klog/v2"
	"strings"

	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
)

const (
	Reason                  = "reason"
	InvolvedObjectKind      = "involvedObject.kind"
	involvedObjectNamespace = "involvedObject.namespace"
)

var SupportField = []string{Reason, involvedObjectNamespace, InvolvedObjectKind}

type EventFilter interface {
	Filter(event *v1.Event) bool
}

type EventTypeFilter struct {
	AllowedTypes []string
}

func NewEventTypeFilter(allowedTypes []string) *EventTypeFilter {
	return &EventTypeFilter{
		AllowedTypes: allowedTypes,
	}
}

func (e *EventTypeFilter) Filter(event *v1.Event) bool {
	for _, allowedType := range e.AllowedTypes {
		if strings.EqualFold(event.Type, allowedType) {
			return true
		}
	}
	return false
}

type EventFieldSelectorFilter struct {
	MatchFieldSelector map[string][]string
}

func NewEventFieldSelectorFilter(fieldSelectors []string) *EventFieldSelectorFilter {
	matchFieldSelector := make(map[string][]string)
	for _, fieldSelector := range fieldSelectors {
		selectors := strings.Split(fieldSelector, ":")
		if len(selectors) != 2 {
			klog.Fatalf("selector format error, selector:%s", fieldSelector)
		}
		field := selectors[0]
		if !slices.Contains(SupportField, field) {
			klog.Fatalf("only support fields: %s, unsupported field:%s, ", SupportField, field)
		}
		match := selectors[1]
		matchStrings := strings.Split(match, "|")
		matchFieldSelector[field] = matchStrings
	}
	return &EventFieldSelectorFilter{
		MatchFieldSelector: matchFieldSelector,
	}
}

func (e *EventFieldSelectorFilter) Filter(event *v1.Event) bool {
	for field := range e.MatchFieldSelector {
		switch field {
		case Reason:
			if e.sliceContains(Reason, event.Reason) {
				return true
			}
		case InvolvedObjectKind:
			if e.sliceContains(InvolvedObjectKind, event.InvolvedObject.Kind) {
				return true
			}
		case involvedObjectNamespace:
			if e.sliceContains(involvedObjectNamespace, event.InvolvedObject.Namespace) {
				return true
			}
		}
	}
	return false
}

func (e *EventFieldSelectorFilter) sliceContains(field, element string) bool {
	selectors := e.MatchFieldSelector[field]
	return slices.Contains(selectors, element)
}
