/* SPDX-License-Identifier: Apache-2.0
 *
 * Copyright 2024 Damian Peckett <damian@pecke.tt>.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package updater provides a way to update Kubernetes resources.
// Inspired by controllerutil, gardener, and the kubernetes statefulset controller.
// https://github.com/gardener/gardener/blob/master/docs/development/kubernetes-clients.md
// https://github.com/kubernetes/kubernetes/blob/ee265c92fec40cd69d1de010b477717e4c142492/pkg/controller/history/controller_history.go#L92
package updater

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MutateFunc func() error

// CreateOrUpdateFromTemplate creates or updates the given object using the given template.
func CreateOrUpdateFromTemplate(ctx context.Context, c client.Client, template client.Object) (client.Object, error) {
	templateHash := HashObject(template)

	obj, ok := template.DeepCopyObject().(client.Object)
	if !ok {
		return nil, fmt.Errorf("expected client object")
	}

	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get object: %w", err)
		}

		if err := StoreHash(obj, templateHash); err != nil {
			return nil, fmt.Errorf("failed to store hash: %w", err)
		}

		if err := c.Create(ctx, obj); err != nil {
			return nil, fmt.Errorf("failed to create object: %w", err)
		}

		if err := c.Get(ctx, key, obj); err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get object: %w", err)
		}

		return obj, nil
	}

	existingHash, err := GetHash(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get hash from object: %w", err)
	}

	if existingHash != templateHash {
		obj = template.DeepCopyObject().(client.Object)

		if err := StoreHash(obj, templateHash); err != nil {
			return nil, fmt.Errorf("failed to store hash: %w", err)
		}

		if err := c.Update(ctx, obj); err != nil {
			return nil, fmt.Errorf("failed to update object: %w", err)
		}

		if err := c.Get(ctx, key, obj); err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get object: %w", err)
		}
	}

	return obj, nil
}

// UpdateStatus updates the status of the given object using a mutating function.
func UpdateStatus(ctx context.Context, c client.Client, obj client.Object, f MutateFunc) error {
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	if f != nil {
		if err := f(); err != nil {
			return fmt.Errorf("failed to mutate object: %w", err)
		}
	}

	if err := c.Status().Update(ctx, obj); err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	return nil
}
