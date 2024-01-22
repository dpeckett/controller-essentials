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

package updater

import (
	"encoding/hex"
	"hash/fnv"
	"io"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/dump"
)

const (
	// AnnotationKey is the key used to store the hash of the template object.
	AnnotationKey = "environment-manager.zymatik.com/template-hash"
)

func GetHash(obj runtime.Object) (string, error) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}

	annotations := metaObj.GetAnnotations()
	if annotations == nil {
		return "", nil
	}

	return annotations[AnnotationKey], nil
}

func StoreHash(obj runtime.Object, hash string) error {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	annotations := metaObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[AnnotationKey] = hash
	metaObj.SetAnnotations(annotations)

	return nil
}

// HashObject returns a hash of the given object.
// This is inspired by the way Kubernetes manages controller revisions in StatefulSets:
// https://github.com/kubernetes/kubernetes/blob/ee265c92fec40cd69d1de010b477717e4c142492/pkg/controller/history/controller_history.go#L92
func HashObject(obj runtime.Object) string {
	h := fnv.New32a()
	_, _ = io.WriteString(h, dump.ForHash(obj))
	return hex.EncodeToString(h.Sum(nil))
}
