package tag_common

import (
	"context"
	"fmt"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"xorm.io/xorm"

	"github.com/lawyer/commons/utils/pager"
	tagcommon "github.com/lawyer/service/tag_common"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

// tagCommonRepo tag repository
type tagCommonRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

// NewTagCommonRepo new repository
func NewTagCommonRepo() tagcommon.TagCommonRepo {
	return &tagCommonRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// GetTagListByIDs get tag list all
func (tr *tagCommonRepo) GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity.Tag, err error) {
	tagList = make([]*entity.Tag, 0)
	session := tr.DB.Context(ctx).In("id", ids)
	session.Where(builder.Eq{"status": entity.TagStatusAvailable})
	err = session.OrderBy("recommend desc,reserved desc,id desc").Find(&tagList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagBySlugName get tag by slug name
func (tr *tagCommonRepo) GetTagBySlugName(ctx context.Context, slugName string) (tagInfo *entity.Tag, exist bool, err error) {
	tagInfo = &entity.Tag{}
	session := tr.DB.Context(ctx).Where("LOWER(slug_name) = ?", slugName)
	session.Where(builder.Eq{"status": entity.TagStatusAvailable})
	exist, err = session.Get(tagInfo)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagListByName get tag list all like name
func (tr *tagCommonRepo) GetTagListByName(ctx context.Context, name string, recommend, reserved bool) (tagList []*entity.Tag, err error) {
	cond := &entity.Tag{}
	session := tr.DB.Context(ctx)
	if len(name) > 0 {
		session.Where("slug_name LIKE ? OR display_name LIKE ?", strings.ToLower(name)+"%", name+"%")
	}
	var columns []string
	if recommend {
		columns = append(columns, "recommend")
		cond.Recommend = true
	}
	if reserved {
		columns = append(columns, "reserved")
		cond.Reserved = true
	}
	if len(columns) > 0 {
		session.UseBool(columns...)
	}
	session.Where(builder.Eq{"status": entity.TagStatusAvailable})

	tagList = make([]*entity.Tag, 0)
	err = session.OrderBy("recommend DESC,reserved DESC,slug_name ASC").Find(&tagList, cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagCommonRepo) GetRecommendTagList(ctx context.Context) (tagList []*entity.Tag, err error) {
	tagList = make([]*entity.Tag, 0)
	cond := &entity.Tag{}
	session := tr.DB.Context(ctx).Where("")
	cond.Recommend = true
	// session.Where(builder.Eq{"status": entity.TagStatusAvailable})
	session.Asc("slug_name")
	session.UseBool("recommend")
	err = session.Find(&tagList, cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagCommonRepo) GetReservedTagList(ctx context.Context) (tagList []*entity.Tag, err error) {
	tagList = make([]*entity.Tag, 0)
	cond := &entity.Tag{}
	session := tr.DB.Context(ctx).Where("")
	cond.Reserved = true
	// session.Where(builder.Eq{"status": entity.TagStatusAvailable})
	session.Asc("slug_name")
	session.UseBool("reserved")
	err = session.Find(&tagList, cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagListByNames get tag list all like name
func (tr *tagCommonRepo) GetTagListByNames(ctx context.Context, names []string) (tagList []*entity.Tag, err error) {
	tagList = make([]*entity.Tag, 0)
	session := tr.DB.Context(ctx).In("slug_name", names).UseBool("recommend", "reserved")
	session.Where(builder.Eq{"status": entity.TagStatusAvailable})
	err = session.OrderBy("recommend desc,reserved desc,id desc").Find(&tagList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagByID get tag one
func (tr *tagCommonRepo) GetTagByID(ctx context.Context, tagID string, includeDeleted bool) (
	tag *entity.Tag, exist bool, err error,
) {
	tag = &entity.Tag{}
	session := tr.DB.Context(ctx).Where(builder.Eq{"id": tagID})
	if !includeDeleted {
		session.Where(builder.Eq{"status": entity.TagStatusAvailable})
	}
	exist, err = session.Get(tag)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagPage get tag page
func (tr *tagCommonRepo) GetTagPage(ctx context.Context, page, pageSize int, tag *entity.Tag, queryCond string) (
	tagList []*entity.Tag, total int64, err error,
) {
	tagList = make([]*entity.Tag, 0)
	session := tr.DB.Context(ctx)

	if len(tag.SlugName) > 0 {
		mainTagCond := builder.And(
			builder.Or(
				builder.Like{"slug_name", fmt.Sprintf("LOWER(%s)", tag.SlugName)},
				builder.Like{"display_name", tag.SlugName},
			),
			builder.Eq{"main_tag_id": 0},
		)
		synonymCond := builder.And(
			builder.Eq{"slug_name": tag.SlugName},
			builder.Neq{"main_tag_id": 0},
		)
		session.Where(builder.Or(mainTagCond, synonymCond))
		tag.SlugName = ""
	} else {
		session.Where(builder.Eq{"main_tag_id": 0})
	}
	session.Where(builder.Eq{"status": entity.TagStatusAvailable})

	switch queryCond {
	case "popular":
		session.Desc("question_count")
	case "name":
		session.Asc("slug_name")
	case "newest":
		session.Desc("created_at")
	}

	total, err = pager.Help(page, pageSize, &tagList, tag, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}

	for i := 0; i < len(tagList); i++ {
		if tagList[i].MainTagID != 0 {
			mainTag, exist, errSynonym := tr.GetTagByID(ctx, strconv.FormatInt(tagList[i].MainTagID, 10), false)
			if errSynonym != nil {
				err = errors.InternalServer(reason.DatabaseError).WithError(errSynonym).WithStack()
				return
			}
			if exist {
				tagList[i] = mainTag
			}
		}
	}

	return
}

// AddTagList add tag
func (tr *tagCommonRepo) AddTagList(ctx context.Context, tagList []*entity.Tag) (err error) {
	addTags := make([]*entity.Tag, 0)
	for _, item := range tagList {
		exist, err := tr.updateDeletedTag(ctx, item)
		if err != nil {
			return err
		}
		if exist {
			continue
		}
		addTags = append(addTags, item)
		item.ID, err = repo.UniqueIDRepo.GenUniqueIDStr(ctx, item.TableName())
		if err != nil {
			return err
		}
		item.RevisionID = "0"
	}
	if len(addTags) == 0 {
		return nil
	}
	_, err = tr.DB.Context(ctx).Insert(addTags)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagCommonRepo) updateDeletedTag(ctx context.Context, tag *entity.Tag) (exist bool, err error) {
	old := &entity.Tag{SlugName: tag.SlugName}
	exist, err = tr.DB.Context(ctx).Get(old)
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist || old.Status != entity.TagStatusDeleted {
		return false, nil
	}
	tag.ID = old.ID
	tag.Status = entity.TagStatusAvailable
	tag.RevisionID = "0"
	if _, err = tr.DB.Context(ctx).ID(tag.ID).Update(tag); err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return true, nil
}

// UpdateTagQuestionCount update tag question count
func (tr *tagCommonRepo) UpdateTagQuestionCount(ctx context.Context, tagID string, questionCount int) (err error) {
	cond := &entity.Tag{QuestionCount: questionCount}
	_, err = tr.DB.Context(ctx).Where(builder.Eq{"id": tagID}).MustCols("question_count").Update(cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagCommonRepo) UpdateTagsAttribute(ctx context.Context, tags []string, attribute string, value bool) (err error) {
	bean := &entity.Tag{}
	switch attribute {
	case "recommend":
		bean.Recommend = value
	case "reserved":
		bean.Reserved = value
	default:
		return
	}
	session := tr.DB.Context(ctx).In("slug_name", tags).Cols(attribute).UseBool(attribute)
	_, err = session.Update(bean)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}