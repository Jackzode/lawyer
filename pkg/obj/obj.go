package obj

import (
	"errors"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/pkg/converter"
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
	return "", errors.New(reason.ObjectNotFound)
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
		return errors.New(reason.ObjectNotFound)
	}
	return nil
}
