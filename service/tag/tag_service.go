package tag

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/permission"

	"github.com/jinzhu/copier"
	"github.com/lawyer/pkg/htmltext"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// TagService user service
type TagService struct {
}

// NewTagService new tag service
func NewTagService() *TagService {
	return &TagService{}
}

// RemoveTag delete tag
func (ts *TagService) RemoveTag(ctx context.Context, req *schema.RemoveTagReq) (err error) {
	//If the tag has associated problems, it cannot be deleted
	tagCount, err := services.TagCommonService.CountTagRelByTagID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if tagCount > 0 {
		return errors.BadRequest(reason.TagIsUsedCannotDelete)
	}

	//If the tag has associated problems, it cannot be deleted
	tagSynonymCount, err := repo.TagRepo.GetTagSynonymCount(ctx, req.TagID)
	if err != nil {
		return err
	}
	if tagSynonymCount > 0 {
		return errors.BadRequest(reason.TagIsUsedCannotDelete)
	}

	// tagRelRepo
	err = repo.TagRepo.RemoveTag(ctx, req.TagID)
	if err != nil {
		return err
	}
	services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         req.TagID,
		OriginalObjectID: req.TagID,
		ActivityTypeKey:  constant.ActTagDeleted,
	})
	return nil
}

// UpdateTag update tag
func (ts *TagService) UpdateTag(ctx context.Context, req *schema.UpdateTagReq) (err error) {
	return services.TagCommonService.UpdateTag(ctx, req)
}

// RecoverTag recover tag
func (ts *TagService) RecoverTag(ctx context.Context, req *schema.RecoverTagReq) (err error) {
	tagInfo, exist, err := repo.TagRepo.MustGetTagByNameOrID(ctx, req.TagID, "")
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.TagNotFound)
	}
	if tagInfo.Status != entity.TagStatusDeleted {
		return nil
	}

	err = repo.TagRepo.RecoverTag(ctx, req.TagID)
	if err != nil {
		return err
	}
	services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		TriggerUserID:    converter.StringToInt64(req.UserID),
		ObjectID:         req.TagID,
		OriginalObjectID: req.TagID,
		ActivityTypeKey:  constant.ActTagUndeleted,
	})
	return nil
}

// GetTagInfo get tag one
func (ts *TagService) GetTagInfo(ctx context.Context, req *schema.GetTagInfoReq) (resp *schema.GetTagResp, err error) {
	var (
		tagInfo *entity.Tag
		exist   bool
	)
	if len(req.ID) > 0 {
		tagInfo, exist, err = services.TagCommonService.GetTagByID(ctx, req.ID)
	} else {
		tagInfo, exist, err = services.TagCommonService.GetTagBySlugName(ctx, req.Name)
	}
	// If user can recover deleted tag, try to search in all tags including deleted tags
	if !exist && req.CanRecover {
		tagInfo, exist, err = repo.TagRepo.MustGetTagByNameOrID(ctx, req.ID, req.Name)
	}
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.TagNotFound)
	}

	resp = &schema.GetTagResp{}
	// if tag is synonyms get original tag info
	if tagInfo.MainTagID > 0 {
		tagInfo, exist, err = services.TagCommonService.GetTagByID(ctx, converter.IntToString(tagInfo.MainTagID))
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.NotFound(reason.TagNotFound)
		}
		resp.MainTagSlugName = tagInfo.SlugName
	}
	resp.TagID = tagInfo.ID
	resp.CreatedAt = tagInfo.CreatedAt.Unix()
	resp.UpdatedAt = tagInfo.UpdatedAt.Unix()
	resp.SlugName = tagInfo.SlugName
	resp.DisplayName = tagInfo.DisplayName
	resp.OriginalText = tagInfo.OriginalText
	resp.ParsedText = tagInfo.ParsedText
	resp.Description = htmltext.FetchExcerpt(tagInfo.ParsedText, "...", 240)
	resp.FollowCount = tagInfo.FollowCount
	resp.QuestionCount = tagInfo.QuestionCount
	resp.Recommend = tagInfo.Recommend
	resp.Reserved = tagInfo.Reserved
	resp.IsFollower = ts.checkTagIsFollow(ctx, req.UserID, tagInfo.ID)
	resp.Status = entity.TagStatusDisplayMapping[tagInfo.Status]
	resp.MemberActions = permission.GetTagPermission(ctx, tagInfo.Status, req.CanEdit, req.CanDelete, req.CanRecover)
	resp.GetExcerpt()
	return resp, nil
}

func (ts *TagService) GetTagsBySlugName(ctx context.Context, tagNames []string) ([]*schema.TagItem, error) {
	tagList := make([]*schema.TagItem, 0)
	tagListInDB, err := services.TagCommonService.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return tagList, err
	}
	for _, tag := range tagListInDB {
		tagItem := &schema.TagItem{}
		copier.Copy(tagItem, tag)
		tagList = append(tagList, tagItem)
	}
	return tagList, nil
}

// GetFollowingTags get following tags
func (ts *TagService) GetFollowingTags(ctx context.Context, userID string) (
	resp []*schema.GetFollowingTagsResp, err error) {
	resp = make([]*schema.GetFollowingTagsResp, 0)
	if len(userID) == 0 {
		return resp, nil
	}
	objIDs, err := repo.FollowRepo.GetFollowIDs(ctx, userID, entity.Tag{}.TableName())
	if err != nil {
		return nil, err
	}
	tagList, err := services.TagCommonService.GetTagListByIDs(ctx, objIDs)
	if err != nil {
		return nil, err
	}
	for _, t := range tagList {
		tagInfo := &schema.GetFollowingTagsResp{
			TagID:       t.ID,
			SlugName:    t.SlugName,
			DisplayName: t.DisplayName,
			Recommend:   t.Recommend,
			Reserved:    t.Reserved,
		}
		if t.MainTagID > 0 {
			mainTag, exist, err := services.TagCommonService.GetTagByID(ctx, converter.IntToString(t.MainTagID))
			if err != nil {
				return nil, err
			}
			if exist {
				tagInfo.MainTagSlugName = mainTag.SlugName
			}
		}
		resp = append(resp, tagInfo)
	}
	return resp, nil
}

// GetTagSynonyms get tag synonyms
func (ts *TagService) GetTagSynonyms(ctx context.Context, req *schema.GetTagSynonymsReq) (
	resp *schema.GetTagSynonymsResp, err error) {
	resp = &schema.GetTagSynonymsResp{Synonyms: make([]*schema.TagSynonym, 0)}
	tag, exist, err := services.TagCommonService.GetTagByID(ctx, req.TagID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.TagNotFound)
	}

	var tagList []*entity.Tag
	var mainTagSlugName string
	if tag.MainTagID > 0 {
		tagList, err = repo.TagRepo.GetTagList(ctx, &entity.Tag{MainTagID: tag.MainTagID})
	} else {
		tagList, err = repo.TagRepo.GetTagList(ctx, &entity.Tag{MainTagID: converter.StringToInt64(tag.ID)})
	}
	if err != nil {
		return
	}

	// get main tag slug name
	if tag.MainTagID > 0 {
		for _, tagInfo := range tagList {
			if tag.MainTagID == 0 {
				mainTagSlugName = tagInfo.SlugName
				break
			}
		}
	} else {
		mainTagSlugName = tag.SlugName
	}

	for _, t := range tagList {
		resp.Synonyms = append(resp.Synonyms, &schema.TagSynonym{
			TagID:           t.ID,
			SlugName:        t.SlugName,
			DisplayName:     t.DisplayName,
			MainTagSlugName: mainTagSlugName,
		})
	}
	resp.MemberActions = permission.GetTagSynonymPermission(ctx, req.CanEdit)
	return
}

// UpdateTagSynonym add tag synonym
func (ts *TagService) UpdateTagSynonym(ctx context.Context, req *schema.UpdateTagSynonymReq) (err error) {
	// format tag slug name
	req.Format()
	addSynonymTagList := make([]string, 0)
	removeSynonymTagList := make([]string, 0)
	mainTagInfo, exist, err := services.TagCommonService.GetTagByID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.TagNotFound)
	}

	// find all exist tag
	for _, item := range req.SynonymTagList {
		if item.SlugName == mainTagInfo.SlugName {
			return errors.BadRequest(reason.TagCannotSetSynonymAsItself)
		}
		addSynonymTagList = append(addSynonymTagList, item.SlugName)
	}
	tagListInDB, err := services.TagCommonService.GetTagListByNames(ctx, addSynonymTagList)
	if err != nil {
		return err
	}
	existTagMapping := make(map[string]*entity.Tag, 0)
	for _, tag := range tagListInDB {
		existTagMapping[tag.SlugName] = tag
	}

	// add tag list
	needAddTagList := make([]*entity.Tag, 0)
	for _, tag := range req.SynonymTagList {
		if existTagMapping[tag.SlugName] != nil {
			continue
		}
		item := &entity.Tag{}
		item.SlugName = tag.SlugName
		item.DisplayName = tag.DisplayName
		item.OriginalText = tag.OriginalText
		item.ParsedText = tag.ParsedText
		item.Status = entity.TagStatusAvailable
		item.UserID = req.UserID
		needAddTagList = append(needAddTagList, item)
	}

	if len(needAddTagList) > 0 {
		err = services.TagCommonService.AddTagList(ctx, needAddTagList)
		if err != nil {
			return err
		}
		// update tag revision
		for _, tag := range needAddTagList {
			existTagMapping[tag.SlugName] = tag
			revisionDTO := &schema.AddRevisionDTO{
				UserID:   req.UserID,
				ObjectID: tag.ID,
				Title:    tag.SlugName,
			}
			tagInfoJson, _ := json.Marshal(tag)
			revisionDTO.Content = string(tagInfoJson)
			revisionID, err := services.RevisionService.AddRevision(ctx, revisionDTO, true)
			if err != nil {
				return err
			}
			services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
				UserID:           req.UserID,
				ObjectID:         tag.ID,
				OriginalObjectID: tag.ID,
				ActivityTypeKey:  constant.ActTagCreated,
				RevisionID:       revisionID,
			})
		}
	}

	// get all old synonyms list
	oldSynonymList, err := repo.TagRepo.GetTagList(ctx, &entity.Tag{MainTagID: converter.StringToInt64(mainTagInfo.ID)})
	if err != nil {
		return err
	}
	for _, oldSynonym := range oldSynonymList {
		if existTagMapping[oldSynonym.SlugName] == nil {
			removeSynonymTagList = append(removeSynonymTagList, oldSynonym.SlugName)
		}
	}

	// remove old synonyms
	if len(removeSynonymTagList) > 0 {
		err = repo.TagRepo.UpdateTagSynonym(ctx, removeSynonymTagList, 0, "")
		if err != nil {
			return err
		}
	}

	// update new synonyms
	if len(addSynonymTagList) > 0 {
		err = repo.TagRepo.UpdateTagSynonym(ctx, addSynonymTagList, converter.StringToInt64(req.TagID), mainTagInfo.SlugName)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetTagWithPage get tag list page
func (ts *TagService) GetTagWithPage(ctx context.Context, req *schema.GetTagWithPageReq) (pageModel *pager.PageModel, err error) {
	tag := &entity.Tag{}
	_ = copier.Copy(tag, req)
	tag.UserID = ""

	page := req.Page
	pageSize := req.PageSize

	tags, total, err := services.TagCommonService.GetTagPage(ctx, page, pageSize, tag, req.QueryCond)
	if err != nil {
		return
	}

	resp := make([]*schema.GetTagPageResp, 0)
	for _, tag := range tags {
		item := &schema.GetTagPageResp{
			TagID:         tag.ID,
			SlugName:      tag.SlugName,
			Description:   htmltext.FetchExcerpt(tag.ParsedText, "...", 240),
			DisplayName:   tag.DisplayName,
			OriginalText:  tag.OriginalText,
			ParsedText:    tag.ParsedText,
			FollowCount:   tag.FollowCount,
			QuestionCount: tag.QuestionCount,
			IsFollower:    ts.checkTagIsFollow(ctx, req.UserID, tag.ID),
			CreatedAt:     tag.CreatedAt.Unix(),
			UpdatedAt:     tag.UpdatedAt.Unix(),
			Recommend:     tag.Recommend,
			Reserved:      tag.Reserved,
		}
		item.GetExcerpt()
		resp = append(resp, item)

	}
	return pager.NewPageModel(total, resp), nil
}

// checkTagIsFollow get tag list page
func (ts *TagService) checkTagIsFollow(ctx context.Context, userID, tagID string) bool {
	if len(userID) == 0 {
		return false
	}
	followed, err := repo.FollowRepo.IsFollowed(ctx, userID, tagID)
	if err != nil {
		log.Error(err)
	}
	return followed
}
