package revision_common

import (
	"context"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"

	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/schema"
)

// RevisionService user service
type RevisionService struct {
}

func NewRevisionService() *RevisionService {
	return &RevisionService{}
}

func (rs *RevisionService) GetUnreviewedRevisionCount(ctx context.Context, req *schema.RevisionSearch) (count int64, err error) {
	if len(req.GetCanReviewObjectTypes()) == 0 {
		return 0, nil
	}
	_, count, err = repo.RevisionRepo.GetUnreviewedRevisionPage(ctx, req.Page, 1, req.GetCanReviewObjectTypes())
	return count, err
}

// AddRevision add revision
//
// autoUpdateRevisionID bool : if autoUpdateRevisionID is true , the object.revision_id will be updated,
// if not need auto update object.revision_id, it must be false.
// example: user can edit the object, but need audit, the revision_id will be updated when admin approved
func (rs *RevisionService) AddRevision(ctx context.Context, req *schema.AddRevisionDTO, autoUpdateRevisionID bool) (
	revisionID string, err error) {
	req.ObjectID = uid.DeShortID(req.ObjectID)
	rev := &entity.Revision{}
	_ = copier.Copy(rev, req)
	err = repo.RevisionRepo.AddRevision(ctx, rev, autoUpdateRevisionID)
	if err != nil {
		return "", err
	}
	return rev.ID, nil
}

// GetRevision get revision
func (rs *RevisionService) GetRevision(ctx context.Context, revisionID string) (
	revision *entity.Revision, err error) {
	revisionInfo, exist, err := repo.RevisionRepo.GetRevisionByID(ctx, revisionID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	return revisionInfo, nil
}

// ExistUnreviewedByObjectID
func (rs *RevisionService) ExistUnreviewedByObjectID(ctx context.Context, objectID string) (revision *entity.Revision, exist bool, err error) {
	objectID = uid.DeShortID(objectID)
	revision, exist, err = repo.RevisionRepo.ExistUnreviewedByObjectID(ctx, objectID)
	return revision, exist, err
}
