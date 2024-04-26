package service

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/repo"
)

// CollectionServicer user service
type CollectionService struct {
}

func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

func (cs *CollectionService) CollectionSwitch(ctx context.Context, req *schema.CollectionSwitchReq) (
	resp *schema.CollectionSwitchResp, err error) {
	collectionGroup, err := repo.CollectionGroupRepo.CreateDefaultGroupIfNotExist(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	collection, exist, err := repo.CollectionRepo.GetOneByObjectIDAndUser(ctx, req.UserID, req.ObjectID)
	if err != nil {
		return nil, err
	}
	if (!req.Bookmark && !exist) || (req.Bookmark && exist) {
		return nil, nil
	}

	if req.Bookmark {
		collection = &entity.Collection{
			UserID:                req.UserID,
			ObjectID:              req.ObjectID,
			UserCollectionGroupID: collectionGroup.ID,
		}
		err = repo.CollectionRepo.AddCollection(ctx, collection)
	} else {
		err = repo.CollectionRepo.RemoveCollection(ctx, collection.ID)
	}
	if err != nil {
		return nil, err
	}

	// For now, we only support bookmark for question, so we just update question collection count
	resp = &schema.CollectionSwitchResp{}
	resp.ObjectCollectionCount, err = QuestionCommonServicer.UpdateCollectionCount(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
