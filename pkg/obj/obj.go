/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package obj

import (
	"github.com/apache/incubator-answer/commons/constant"
	"github.com/apache/incubator-answer/commons/constant/reason"
	"github.com/apache/incubator-answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
)

// GetObjectTypeStrByObjectID get object key by object id
func GetObjectTypeStrByObjectID(objectID string) (objectTypeStr string, err error) {
	if err = checkObjectID(objectID); err != nil {
		return "", err
	}
	objectTypeNumber := converter.StringToInt(objectID[1:4])
	objectTypeStr, ok := constant.ObjectTypeNumberMapping[objectTypeNumber]
	if ok {
		return objectTypeStr, nil
	}
	return "", errors.BadRequest(reason.ObjectNotFound)
}

// GetObjectTypeNumberByObjectID get object type by object id
func GetObjectTypeNumberByObjectID(objectID string) (objectTypeNumber int, err error) {
	if err := checkObjectID(objectID); err != nil {
		return 0, err
	}
	return converter.StringToInt(objectID[1:4]), nil
}

func checkObjectID(objectID string) (err error) {
	if len(objectID) < 5 {
		return errors.BadRequest(reason.ObjectNotFound)
	}
	return nil
}