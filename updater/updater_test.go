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

package updater_test

import (
	"context"
	"testing"

	"github.com/dpeckett/controller-essentials/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateOrUpdateFromTemplate(t *testing.T) {
	scheme := runtime.NewScheme()

	err := appsv1.AddToScheme(scheme)
	require.NoError(t, err)

	template := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	ctx := context.Background()

	obj, err := updater.CreateOrUpdateFromTemplate(ctx, c, &template)
	require.NoError(t, err)

	hash, err := updater.GetHash(obj)
	require.NoError(t, err)

	assert.Equal(t, "275e0e96", hash)

	updatedTemplate := template.DeepCopy()
	updatedTemplate.Spec.Replicas = ptr.To(int32(2))

	obj, err = updater.CreateOrUpdateFromTemplate(ctx, c, updatedTemplate)
	require.NoError(t, err)

	hash, err = updater.GetHash(obj)
	require.NoError(t, err)

	assert.Equal(t, "c2f1de77", hash)
}
