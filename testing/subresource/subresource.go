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

// Package subresource provides a fake client for testing subresources.
package subresource

import (
	"context"
	"encoding/json"
	"sync"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/jinzhu/copier"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a fake subresource client.s
type Client struct {
	mu           sync.Mutex
	objects      map[client.ObjectKey]runtime.Object
	codecFactory serializer.CodecFactory
}

func NewClient(scheme *runtime.Scheme) *Client {
	return &Client{
		objects:      make(map[client.ObjectKey]runtime.Object),
		codecFactory: serializer.NewCodecFactory(scheme),
	}
}

// Reset clears the contents of the client.
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.objects = make(map[client.ObjectKey]runtime.Object)
}

func (c *Client) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := client.ObjectKeyFromObject(obj)
	if obj, ok := c.objects[key]; ok {
		return copier.Copy(subResource, obj)
	}

	return apierrors.NewNotFound(schema.GroupResource{}, key.Name)
}

func (c *Client) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := client.ObjectKeyFromObject(obj)
	c.objects[key] = obj.DeepCopyObject()

	return nil
}

func (c *Client) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := client.ObjectKeyFromObject(obj)
	c.objects[key] = obj.DeepCopyObject()

	return nil
}

func (c *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	patchJSON, err := patch.Data(obj)
	if err != nil {
		return err
	}

	objJSON, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	patchedJSON, err := jsonpatch.MergePatch(objJSON, patchJSON)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(patchedJSON, obj); err != nil {
		return err
	}

	key := client.ObjectKeyFromObject(obj)
	c.objects[key] = obj.DeepCopyObject()

	return nil
}
