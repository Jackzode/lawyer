package follow

import (
	"context"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/repo"
)

type FollowRepo interface {
	Follow(ctx context.Context, objectId, userId string) error
	FollowCancel(ctx context.Context, objectId, userId string) error
}

type FollowService struct {
}

func NewFollowService() *FollowService {
	return &FollowService{}
}

// Follow or cancel follow object
func (fs *FollowService) Follow(ctx context.Context, dto *schema.FollowDTO) (resp schema.FollowResp, err error) {
	if dto.IsCancel {
		err = repo.FollowFollowRepo.FollowCancel(ctx, dto.ObjectID, dto.UserID)
	} else {
		err = repo.FollowFollowRepo.Follow(ctx, dto.ObjectID, dto.UserID)
	}
	if err != nil {
		return resp, err
	}
	follows, err := repo.FollowRepo.GetFollowAmount(ctx, dto.ObjectID)
	if err != nil {
		return resp, err
	}

	resp.Follows = follows
	resp.IsFollowed = !dto.IsCancel
	return resp, nil
}

// UpdateFollowTags update user follow tags
func (fs *FollowService) UpdateFollowTags(ctx context.Context, req *schema.UpdateFollowTagsReq) (err error) {
	objIDs, err := repo.FollowRepo.GetFollowIDs(ctx, req.UserID, entity.Tag{}.TableName())
	if err != nil {
		return
	}
	oldFollowTagList, err := repo.TagRepo.GetTagListByIDs(ctx, objIDs)
	if err != nil {
		return err
	}
	oldTagMapping := make(map[string]bool)
	for _, tag := range oldFollowTagList {
		oldTagMapping[tag.SlugName] = true
	}

	newTagMapping := make(map[string]bool)
	for _, tag := range req.SlugNameList {
		newTagMapping[tag] = true
	}

	// cancel follow
	for _, tag := range oldFollowTagList {
		if !newTagMapping[tag.SlugName] {
			err := repo.FollowFollowRepo.FollowCancel(ctx, tag.ID, req.UserID)
			if err != nil {
				return err
			}
		}
	}

	// new follow
	for _, tagSlugName := range req.SlugNameList {
		if !oldTagMapping[tagSlugName] {
			tagInfo, exist, err := repo.TagRepo.GetTagBySlugName(ctx, tagSlugName)
			if err != nil {
				return err
			}
			if !exist {
				continue
			}
			err = repo.FollowFollowRepo.Follow(ctx, tagInfo.ID, req.UserID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
