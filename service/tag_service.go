package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	glog "github.com/lawyer/commons/logger"
	"github.com/lawyer/commons/site"
	"github.com/lawyer/commons/utils/pager"
	"github.com/lawyer/pkg/htmltext"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/permission"
	"sort"
	"strings"

	"errors"
	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
)

type TagRepo interface {
	AddTagList(ctx context.Context, tagList []*entity.Tag) (err error)
	GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity.Tag, err error)
	GetTagBySlugName(ctx context.Context, slugName string) (tagInfo *entity.Tag, exist bool, err error)
	GetTagListByName(ctx context.Context, name string, recommend, reserved bool) (tagList []*entity.Tag, err error)
	GetTagListByNames(ctx context.Context, names []string) (tagList []*entity.Tag, err error)
	GetTagByID(ctx context.Context, tagID string, includeDeleted bool) (tag *entity.Tag, exist bool, err error)
	GetTagPage(ctx context.Context, page, pageSize int, tag *entity.Tag, queryCond string) (tagList []*entity.Tag, total int64, err error)
	GetRecommendTagList(ctx context.Context) (tagList []*entity.Tag, err error)
	GetReservedTagList(ctx context.Context) (tagList []*entity.Tag, err error)
	UpdateTagsAttribute(ctx context.Context, tags []string, attribute string, value bool) (err error)
	UpdateTagQuestionCount(ctx context.Context, tagID string, questionCount int) (err error)
	RemoveTag(ctx context.Context, tagID string) (err error)
	UpdateTag(ctx context.Context, tag *entity.Tag) (err error)
	RecoverTag(ctx context.Context, tagID string) (err error)
	MustGetTagByNameOrID(ctx context.Context, tagID, slugName string) (tag *entity.Tag, exist bool, err error)
	UpdateTagSynonym(ctx context.Context, tagSlugNameList []string, mainTagID int64, mainTagSlugName string) (err error)
	GetTagSynonymCount(ctx context.Context, tagID string) (count int64, err error)
	GetTagList(ctx context.Context, tag *entity.Tag) (tagList []*entity.Tag, err error)
}

type TagRelRepo interface {
	AddTagRelList(ctx context.Context, tagList []*entity.TagRel) (err error)
	RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	HideTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	RemoveTagRelListByIDs(ctx context.Context, ids []int64) (err error)
	EnableTagRelByIDs(ctx context.Context, ids []int64) (err error)
	GetObjectTagRelWithoutStatus(ctx context.Context, objectId, tagID string) (tagRel *entity.TagRel, exist bool, err error)
	GetObjectTagRelList(ctx context.Context, objectId string) (tagListList []*entity.TagRel, err error)
	BatchGetObjectTagRelList(ctx context.Context, objectIds []string) (tagListList []*entity.TagRel, err error)
	CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error)
}

// TagServicer user service
type TagService struct {
}

// NewTagCommonService new tag service
func NewTagService() *TagService {
	return &TagService{}
}

// RemoveTag delete tag
func (ts *TagService) RemoveTag(ctx context.Context, req *schema.RemoveTagReq) (err error) {
	//If the tag has associated problems, it cannot be deleted
	tagCount, err := ts.CountTagRelByTagID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if tagCount > 0 {
		return errors.New(reason.TagIsUsedCannotDelete)
	}

	//If the tag has associated problems, it cannot be deleted
	tagSynonymCount, err := repo.TagRepo.GetTagSynonymCount(ctx, req.TagID)
	if err != nil {
		return err
	}
	if tagSynonymCount > 0 {
		return errors.New(reason.TagIsUsedCannotDelete)
	}

	// tagRelRepo
	err = repo.TagRepo.RemoveTag(ctx, req.TagID)
	if err != nil {
		return err
	}
	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         req.TagID,
		OriginalObjectID: req.TagID,
		ActivityTypeKey:  constant.ActTagDeleted,
	})
	return nil
}

// UpdateTag update tag
//func (ts *TagService) UpdateTag(ctx context.Context, req *schema.UpdateTagReq) (err error) {
//	return TagCommonServicer.UpdateTag(ctx, req)
//}

// RecoverTag recover tag
func (ts *TagService) RecoverTag(ctx context.Context, req *schema.RecoverTagReq) (err error) {
	tagInfo, exist, err := repo.TagRepo.MustGetTagByNameOrID(ctx, req.TagID, "")
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(reason.TagNotFound)
	}
	if tagInfo.Status != entity.TagStatusDeleted {
		return nil
	}

	err = repo.TagRepo.RecoverTag(ctx, req.TagID)
	if err != nil {
		return err
	}
	ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
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
		tagInfo, exist, err = ts.GetTagByID(ctx, req.ID)
	} else {
		tagInfo, exist, err = ts.GetTagBySlugName(ctx, req.Name)
	}
	// If user can recover deleted tag, try to search in all tags including deleted tags
	if !exist && req.CanRecover {
		tagInfo, exist, err = repo.TagRepo.MustGetTagByNameOrID(ctx, req.ID, req.Name)
	}
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.New(reason.TagNotFound)
	}

	resp = &schema.GetTagResp{}
	// if tag is synonyms get original tag info
	if tagInfo.MainTagID > 0 {
		tagInfo, exist, err = ts.GetTagByID(ctx, converter.IntToString(tagInfo.MainTagID))
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.New(reason.TagNotFound)
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
	tagListInDB, err := ts.GetTagListByNames(ctx, tagNames)
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
	tagList, err := ts.GetTagListByIDs(ctx, objIDs)
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
			mainTag, exist, err := ts.GetTagByID(ctx, converter.IntToString(t.MainTagID))
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
	tag, exist, err := ts.GetTagByID(ctx, req.TagID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.New(reason.TagNotFound)
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
	mainTagInfo, exist, err := ts.GetTagByID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(reason.TagNotFound)
	}

	// find all exist tag
	for _, item := range req.SynonymTagList {
		if item.SlugName == mainTagInfo.SlugName {
			return errors.New(reason.TagCannotSetSynonymAsItself)
		}
		addSynonymTagList = append(addSynonymTagList, item.SlugName)
	}
	tagListInDB, err := ts.GetTagListByNames(ctx, addSynonymTagList)
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
		err = ts.AddTagList(ctx, needAddTagList)
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
			revisionID, err := RevisionComServicer.AddRevision(ctx, revisionDTO, true)
			if err != nil {
				return err
			}
			ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
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
	//
	_ = copier.Copy(tag, req)
	tag.UserID = ""

	page := req.Page
	pageSize := req.PageSize

	tags, total, err := ts.GetTagPage(ctx, page, pageSize, tag, req.QueryCond)
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
		glog.Slog.Error(err)
	}
	return followed
}

// SearchTagLike get tag list all
func (ts *TagService) SearchTagLike(ctx context.Context, req *schema.SearchTagLikeReq) (resp []schema.SearchTagLikeResp, err error) {
	tags, err := repo.TagRepo.GetTagListByName(ctx, req.Tag, len(req.Tag) == 0, false)
	if err != nil {
		return
	}
	ts.TagsFormatRecommendAndReserved(ctx, tags)
	mainTagId := make([]string, 0)
	for _, tag := range tags {
		if tag.MainTagID != 0 {
			mainTagId = append(mainTagId, converter.IntToString(tag.MainTagID))
		}
	}
	mainTagMap := make(map[string]*entity.Tag)
	if len(mainTagId) > 0 {
		mainTagList, err := repo.TagRepo.GetTagListByIDs(ctx, mainTagId)
		if err != nil {
			return nil, err
		}
		for _, tag := range mainTagList {
			mainTagMap[tag.ID] = tag
		}
	}
	for _, tag := range tags {
		if tag.MainTagID == 0 {
			continue
		}
		mainTagID := converter.IntToString(tag.MainTagID)
		if _, ok := mainTagMap[mainTagID]; ok {
			tag.SlugName = mainTagMap[mainTagID].SlugName
			tag.DisplayName = mainTagMap[mainTagID].DisplayName
			tag.Reserved = mainTagMap[mainTagID].Reserved
			tag.Recommend = mainTagMap[mainTagID].Recommend
		}
	}
	resp = make([]schema.SearchTagLikeResp, 0)
	repetitiveTag := make(map[string]bool)
	for _, tag := range tags {
		if _, ok := repetitiveTag[tag.SlugName]; !ok {
			item := schema.SearchTagLikeResp{}
			item.SlugName = tag.SlugName
			item.DisplayName = tag.DisplayName
			item.Recommend = tag.Recommend
			item.Reserved = tag.Reserved
			resp = append(resp, item)
			repetitiveTag[tag.SlugName] = true
		}
	}
	return resp, nil
}

func (ts *TagService) GetSiteWriteRecommendTag(ctx context.Context) (tags []string, err error) {
	tags = make([]string, 0)
	list, err := repo.TagRepo.GetRecommendTagList(ctx)
	if err != nil {
		return tags, err
	}
	for _, item := range list {
		tags = append(tags, item.SlugName)
	}
	return tags, nil
}

func (ts *TagService) SetSiteWriteTag(ctx context.Context, recommendTags, reservedTags []string, userID string) (
	errFields []*validator.FormErrorField, err error) {
	recommendErr := ts.CheckTag(ctx, recommendTags, userID)
	reservedErr := ts.CheckTag(ctx, reservedTags, userID)
	if recommendErr != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "recommend_tags",
			ErrorMsg:   recommendErr.Error(),
		})
		err = recommendErr
	}
	if reservedErr != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "reserved_tags",
			ErrorMsg:   reservedErr.Error(),
		})
		err = reservedErr
	}
	if len(errFields) > 0 {
		return errFields, err
	}

	err = ts.SetTagsAttribute(ctx, recommendTags, "recommend")
	if err != nil {
		return nil, err
	}
	err = ts.SetTagsAttribute(ctx, reservedTags, "reserved")
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (ts *TagService) GetSiteWriteReservedTag(ctx context.Context) (tags []string, err error) {
	tags = make([]string, 0)
	list, err := repo.TagRepo.GetReservedTagList(ctx)
	if err != nil {
		return tags, err
	}
	for _, item := range list {
		tags = append(tags, item.SlugName)
	}
	return tags, nil
}

// SetTagsAttribute
func (ts *TagService) SetTagsAttribute(ctx context.Context, tags []string, attribute string) (err error) {
	var tagslist []string
	switch attribute {
	case "recommend":
		tagslist, err = ts.GetSiteWriteRecommendTag(ctx)
	case "reserved":
		tagslist, err = ts.GetSiteWriteReservedTag(ctx)
	default:
		return
	}
	if err != nil {
		return err
	}
	err = repo.TagRepo.UpdateTagsAttribute(ctx, tagslist, attribute, false)
	if err != nil {
		return err
	}
	err = repo.TagRepo.UpdateTagsAttribute(ctx, tags, attribute, true)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TagService) GetTagListByNames(ctx context.Context, tagNames []string) ([]*entity.Tag, error) {
	for k, tagname := range tagNames {
		tagNames[k] = strings.ToLower(tagname)
	}
	tagList, err := repo.TagRepo.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return nil, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return tagList, nil
}

func (ts *TagService) ExistRecommend(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	taginfo := site.Config.GetSiteWrite()
	if !taginfo.RequiredTag {
		return true, nil
	}
	tagNames := make([]string, 0)
	for _, item := range tags {
		item.SlugName = strings.ReplaceAll(item.SlugName, " ", "-")
		tagNames = append(tagNames, item.SlugName)
	}
	list, err := ts.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return false, err
	}
	for _, item := range list {
		if item.Recommend {
			return true, nil
		}
	}
	return false, nil
}

func (ts *TagService) HasNewTag(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	tagNames := make([]string, 0)
	tagMap := make(map[string]bool)
	for _, item := range tags {
		item.SlugName = strings.ReplaceAll(item.SlugName, " ", "-")
		tagNames = append(tagNames, item.SlugName)
		tagMap[item.SlugName] = false
	}
	list, err := ts.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return true, err
	}
	for _, item := range list {
		_, ok := tagMap[item.SlugName]
		if ok {
			tagMap[item.SlugName] = true
		}
	}
	for _, has := range tagMap {
		if !has {
			return true, nil
		}
	}
	return false, nil
}

// GetObjectTag get object tag
func (ts *TagService) GetObjectTag(ctx context.Context, objectId string) (objTags []*schema.TagResp, err error) {
	tagsInfoList, err := ts.GetObjectEntityTag(ctx, objectId)
	if err != nil {
		return nil, err
	}
	return schema.TagFormat(tagsInfoList)
}

// AddTag get object tag
func (ts *TagService) AddTag(ctx context.Context, req *schema.AddTagReq) (resp *schema.AddTagResp, err error) {
	_, exist, err := ts.GetTagBySlugName(ctx, req.SlugName)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New(reason.TagAlreadyExist)
	}
	SlugName := strings.ReplaceAll(req.SlugName, " ", "-")
	SlugName = strings.ToLower(SlugName)
	tagInfo := &entity.Tag{
		SlugName:     SlugName,
		DisplayName:  req.DisplayName,
		OriginalText: req.OriginalText,
		ParsedText:   req.ParsedText,
		Status:       entity.TagStatusAvailable,
		UserID:       req.UserID,
	}
	tagList := []*entity.Tag{tagInfo}
	err = repo.TagRepo.AddTagList(ctx, tagList)
	if err != nil {
		return nil, err
	}
	revisionDTO := &schema.AddRevisionDTO{
		UserID:   req.UserID,
		ObjectID: tagInfo.ID,
		Title:    tagInfo.SlugName,
	}
	tagInfoJson, _ := json.Marshal(tagInfo)
	revisionDTO.Content = string(tagInfoJson)
	_, err = RevisionComServicer.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return nil, err
	}
	return &schema.AddTagResp{SlugName: tagInfo.SlugName}, nil
}

// AddTagList get object tag
func (ts *TagService) AddTagList(ctx context.Context, tagList []*entity.Tag) (err error) {
	return repo.TagRepo.AddTagList(ctx, tagList)
}

// GetTagByID get object tag
func (ts *TagService) GetTagByID(ctx context.Context, tagID string) (tag *entity.Tag, exist bool, err error) {
	tag, exist, err = repo.TagRepo.GetTagByID(ctx, tagID, false)
	if !exist {
		return
	}
	ts.tagFormatRecommendAndReserved(ctx, tag)
	return
}

// GetTagBySlugName get object tag
func (ts *TagService) GetTagBySlugName(ctx context.Context, slugName string) (tag *entity.Tag, exist bool, err error) {
	tag, exist, err = repo.TagRepo.GetTagBySlugName(ctx, slugName)
	if !exist {
		return
	}
	ts.tagFormatRecommendAndReserved(ctx, tag)
	return
}

// GetTagListByIDs get object tag
func (ts *TagService) GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity.Tag, err error) {
	tagList, err = repo.TagRepo.GetTagListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return
}

// GetTagPage get object tag
func (ts *TagService) GetTagPage(ctx context.Context, page, pageSize int, tag *entity.Tag, queryCond string) (
	tagList []*entity.Tag, total int64, err error) {
	tagList, total, err = repo.TagRepo.GetTagPage(ctx, page, pageSize, tag, queryCond)
	if err != nil {
		return nil, 0, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return
}

func (ts *TagService) GetObjectEntityTag(ctx context.Context, objectId string) (objTags []*entity.Tag, err error) {
	tagIDList := make([]string, 0)
	tagList, err := repo.TagRelRepo.GetObjectTagRelList(ctx, objectId)
	if err != nil {
		return nil, err
	}
	for _, tag := range tagList {
		tagIDList = append(tagIDList, tag.TagID)
	}
	objTags, err = ts.GetTagListByIDs(ctx, tagIDList)
	if err != nil {
		return nil, err
	}
	return objTags, nil
}

func (ts *TagService) TagsFormatRecommendAndReserved(ctx context.Context, tagList []*entity.Tag) {
	if len(tagList) == 0 {
		return
	}
	tagConfig := site.Config.GetSiteWrite()
	if !tagConfig.RequiredTag {
		for _, tag := range tagList {
			tag.Recommend = false
		}
	}
}

func (ts *TagService) tagFormatRecommendAndReserved(ctx context.Context, tag *entity.Tag) {
	if tag == nil {
		return
	}
	tagConfig := site.Config.GetSiteWrite()
	if !tagConfig.RequiredTag {
		tag.Recommend = false
	}
}

// BatchGetObjectTag batch get object tag
func (ts *TagService) BatchGetObjectTag(ctx context.Context, objectIds []string) (map[string][]*schema.TagResp, error) {
	objectIDTagMap := make(map[string][]*schema.TagResp)
	if len(objectIds) == 0 {
		return objectIDTagMap, nil
	}
	objectTagRelList, err := repo.TagRelRepo.BatchGetObjectTagRelList(ctx, objectIds)
	if err != nil {
		return objectIDTagMap, err
	}
	tagIDList := make([]string, 0)
	for _, tag := range objectTagRelList {
		tagIDList = append(tagIDList, tag.TagID)
	}
	tagsInfoList, err := ts.GetTagListByIDs(ctx, tagIDList)
	if err != nil {
		return objectIDTagMap, err
	}
	tagsInfoMapping := make(map[string]*entity.Tag)
	tagsRank := make(map[string]int) // Used for sorting
	for idx, item := range tagsInfoList {
		tagsInfoMapping[item.ID] = item
		tagsRank[item.ID] = idx
	}

	for _, item := range objectTagRelList {
		_, ok := tagsInfoMapping[item.TagID]
		if ok {
			tagInfo := tagsInfoMapping[item.TagID]
			t := &schema.TagResp{
				ID:              tagInfo.ID,
				SlugName:        tagInfo.SlugName,
				DisplayName:     tagInfo.DisplayName,
				MainTagSlugName: tagInfo.MainTagSlugName,
				Recommend:       tagInfo.Recommend,
				Reserved:        tagInfo.Reserved,
			}
			objectIDTagMap[item.ObjectID] = append(objectIDTagMap[item.ObjectID], t)
		}
	}
	// The sorting in tagsRank is correct, object tags should be sorted by tagsRank
	for _, objectTags := range objectIDTagMap {
		sort.SliceStable(objectTags, func(i, j int) bool {
			return tagsRank[objectTags[i].ID] < tagsRank[objectTags[j].ID]
		})
	}
	return objectIDTagMap, nil
}

func (ts *TagService) CheckTag(ctx context.Context, tags []string, userID string) (err error) {
	if len(tags) == 0 {
		return nil
	}

	// find tags name
	tagListInDb, err := ts.GetTagListByNames(ctx, tags)
	if err != nil {
		return err
	}

	tagInDbMapping := make(map[string]*entity.Tag)
	checktags := make([]string, 0)

	for _, tag := range tagListInDb {
		if tag.MainTagID != 0 {
			checktags = append(checktags, fmt.Sprintf("\"%s\"", tag.SlugName))
		}
		tagInDbMapping[tag.SlugName] = tag
	}
	if len(checktags) > 0 {
		err = errors.New(reason.TagNotContainSynonym)
		return err
	}

	addTagList := make([]*entity.Tag, 0)
	addTagMsgList := make([]string, 0)
	for _, tag := range tags {
		_, ok := tagInDbMapping[tag]
		if ok {
			continue
		}
		item := &entity.Tag{}
		item.SlugName = tag
		item.DisplayName = tag
		item.OriginalText = ""
		item.ParsedText = ""
		item.Status = entity.TagStatusAvailable
		item.UserID = userID
		addTagList = append(addTagList, item)
		addTagMsgList = append(addTagMsgList, tag)
	}

	if len(addTagList) > 0 {
		err = errors.New(reason.TagNotFound)
		return err

	}

	return nil
}

// CheckTagsIsChange
func (ts *TagService) CheckTagsIsChange(ctx context.Context, tagNameList, oldtagNameList []string) bool {
	check := make(map[string]bool)
	if len(tagNameList) != len(oldtagNameList) {
		return true
	}
	for _, item := range tagNameList {
		check[item] = false
	}
	for _, item := range oldtagNameList {
		_, ok := check[item]
		if !ok {
			return true
		}
		check[item] = true
	}
	for _, value := range check {
		if !value {
			return true
		}
	}
	return false
}

func (ts *TagService) CheckChangeReservedTag(ctx context.Context, oldobjectTagData, objectTagData []*entity.Tag) (bool, bool, []string, []string) {
	reservedTagsMap := make(map[string]bool)
	needTagsMap := make([]string, 0)
	notNeedTagsMap := make([]string, 0)
	for _, tag := range objectTagData {
		if tag.Reserved {
			reservedTagsMap[tag.SlugName] = true
		}
	}
	for _, tag := range oldobjectTagData {
		if tag.Reserved {
			_, ok := reservedTagsMap[tag.SlugName]
			if !ok {
				needTagsMap = append(needTagsMap, tag.SlugName)
			} else {
				reservedTagsMap[tag.SlugName] = false
			}
		}
	}

	for k, v := range reservedTagsMap {
		if v {
			notNeedTagsMap = append(notNeedTagsMap, k)
		}
	}

	if len(needTagsMap) > 0 {
		return false, true, needTagsMap, []string{}
	}

	if len(notNeedTagsMap) > 0 {
		return true, false, []string{}, notNeedTagsMap
	}

	return true, true, []string{}, []string{}
}

// ObjectChangeTag change object tag list
func (ts *TagService) ObjectChangeTag(ctx context.Context, objectTagData *schema.TagChange) (err error) {
	if len(objectTagData.Tags) == 0 {
		return nil
	}

	thisObjTagNameList := make([]string, 0)
	thisObjTagIDList := make([]string, 0)
	for _, t := range objectTagData.Tags {
		t.SlugName = strings.ToLower(t.SlugName)
		thisObjTagNameList = append(thisObjTagNameList, t.SlugName)
	}

	// find tags name
	tagListInDb, err := repo.TagRepo.GetTagListByNames(ctx, thisObjTagNameList)
	if err != nil {
		return err
	}

	tagInDbMapping := make(map[string]*entity.Tag)
	for _, tag := range tagListInDb {
		tagInDbMapping[strings.ToLower(tag.SlugName)] = tag
		thisObjTagIDList = append(thisObjTagIDList, tag.ID)
	}

	addTagList := make([]*entity.Tag, 0)
	for _, tag := range objectTagData.Tags {
		_, ok := tagInDbMapping[strings.ToLower(tag.SlugName)]
		if ok {
			continue
		}
		item := &entity.Tag{}
		item.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		item.DisplayName = tag.DisplayName
		item.OriginalText = tag.OriginalText
		item.ParsedText = tag.ParsedText
		item.Status = entity.TagStatusAvailable
		item.UserID = objectTagData.UserID
		addTagList = append(addTagList, item)
	}

	if len(addTagList) > 0 {
		err = repo.TagRepo.AddTagList(ctx, addTagList)
		if err != nil {
			return err
		}
		for _, tag := range addTagList {
			thisObjTagIDList = append(thisObjTagIDList, tag.ID)
			revisionDTO := &schema.AddRevisionDTO{
				UserID:   objectTagData.UserID,
				ObjectID: tag.ID,
				Title:    tag.SlugName,
			}
			tagInfoJson, _ := json.Marshal(tag)
			revisionDTO.Content = string(tagInfoJson)
			revisionID, err := RevisionComServicer.AddRevision(ctx, revisionDTO, true)
			if err != nil {
				return err
			}
			ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
				UserID:           objectTagData.UserID,
				ObjectID:         tag.ID,
				OriginalObjectID: tag.ID,
				ActivityTypeKey:  constant.ActTagCreated,
				RevisionID:       revisionID,
			})
		}
	}

	err = ts.CreateOrUpdateTagRelList(ctx, objectTagData.ObjectID, thisObjTagIDList)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TagService) CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error) {
	return repo.TagRelRepo.CountTagRelByTagID(ctx, tagID)
}

// RefreshTagQuestionCount refresh tag question count
func (ts *TagService) RefreshTagQuestionCount(ctx context.Context, tagIDs []string) (err error) {
	for _, tagID := range tagIDs {
		count, err := repo.TagRelRepo.CountTagRelByTagID(ctx, tagID)
		if err != nil {
			return err
		}
		err = repo.TagRepo.UpdateTagQuestionCount(ctx, tagID, int(count))
		if err != nil {
			return err
		}
		glog.Slog.Debugf("tag count updated %s %d", tagID, count)
	}
	return nil
}

func (ts *TagService) RefreshTagCountByQuestionID(ctx context.Context, questionID string) (err error) {
	tagListList, err := repo.TagRelRepo.GetObjectTagRelList(ctx, questionID)
	if err != nil {
		return err
	}
	tagIDs := make([]string, 0)
	for _, item := range tagListList {
		tagIDs = append(tagIDs, item.TagID)
	}
	err = ts.RefreshTagQuestionCount(ctx, tagIDs)
	if err != nil {
		return err
	}
	return nil
}

// RemoveTagRelListByObjectID remove tag relation by object id
func (ts *TagService) RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.RemoveTagRelListByObjectID(ctx, objectID)
}

// RecoverTagRelListByObjectID recover tag relation by object id
func (ts *TagService) RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.RecoverTagRelListByObjectID(ctx, objectID)
}

func (ts *TagService) HideTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.HideTagRelListByObjectID(ctx, objectID)
}

func (ts *TagService) ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.ShowTagRelListByObjectID(ctx, objectID)
}

// CreateOrUpdateTagRelList if tag relation is exists update status, if not create it
func (ts *TagService) CreateOrUpdateTagRelList(ctx context.Context, objectId string, tagIDs []string) (err error) {
	addTagIDMapping := make(map[string]bool)
	needRefreshTagIDs := make([]string, 0)
	for _, t := range tagIDs {
		addTagIDMapping[t] = true
	}

	// get all old relation
	oldTagRelList, err := repo.TagRelRepo.GetObjectTagRelList(ctx, objectId)
	if err != nil {
		return err
	}
	var deleteTagRel []int64
	for _, rel := range oldTagRelList {
		if !addTagIDMapping[rel.TagID] {
			deleteTagRel = append(deleteTagRel, rel.ID)
			needRefreshTagIDs = append(needRefreshTagIDs, rel.TagID)
		}
	}

	addTagRelList := make([]*entity.TagRel, 0)
	enableTagRelList := make([]int64, 0)
	for _, tagID := range tagIDs {
		needRefreshTagIDs = append(needRefreshTagIDs, tagID)
		rel, exist, err := repo.TagRelRepo.GetObjectTagRelWithoutStatus(ctx, objectId, tagID)
		if err != nil {
			return err
		}
		// if not exist add tag relation
		if !exist {
			addTagRelList = append(addTagRelList, &entity.TagRel{
				TagID: tagID, ObjectID: objectId, Status: entity.TagStatusAvailable,
			})
		}
		// if exist and has been removed, that should be enabled
		if exist && rel.Status != entity.TagStatusAvailable {
			enableTagRelList = append(enableTagRelList, rel.ID)
		}
	}

	if len(deleteTagRel) > 0 {
		if err = repo.TagRelRepo.RemoveTagRelListByIDs(ctx, deleteTagRel); err != nil {
			return err
		}
	}
	if len(addTagRelList) > 0 {
		if err = repo.TagRelRepo.AddTagRelList(ctx, addTagRelList); err != nil {
			return err
		}
	}
	if len(enableTagRelList) > 0 {
		if err = repo.TagRelRepo.EnableTagRelByIDs(ctx, enableTagRelList); err != nil {
			return err
		}
	}

	err = ts.RefreshTagQuestionCount(ctx, needRefreshTagIDs)
	if err != nil {
		glog.Slog.Error(err)
	}
	return nil
}

func (ts *TagService) UpdateTag(ctx context.Context, req *schema.UpdateTagReq) (err error) {
	var canUpdate bool
	_, existUnreviewed, err := RevisionComServicer.ExistUnreviewedByObjectID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if existUnreviewed {
		err = errors.New(reason.AnswerCannotUpdate)
		return err
	}

	tagInfo, exist, err := ts.GetTagByID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(reason.TagNotFound)
	}
	//If the content is the same, ignore it
	if tagInfo.OriginalText == req.OriginalText &&
		tagInfo.DisplayName == req.DisplayName &&
		tagInfo.SlugName == req.SlugName {
		return nil
	}

	tagInfo.SlugName = req.SlugName
	tagInfo.DisplayName = req.DisplayName
	tagInfo.OriginalText = req.OriginalText
	tagInfo.ParsedText = req.ParsedText

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   req.UserID,
		ObjectID: tagInfo.ID,
		Title:    tagInfo.SlugName,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview {
		canUpdate = true
		err = repo.TagRepo.UpdateTag(ctx, tagInfo)
		if err != nil {
			return err
		}
		if tagInfo.MainTagID == 0 && len(req.SlugName) > 0 {
			glog.Slog.Debugf("tag %s update slug_name", tagInfo.SlugName)
			tagList, err := repo.TagRepo.GetTagList(ctx, &entity.Tag{MainTagID: converter.StringToInt64(tagInfo.ID)})
			if err != nil {
				return err
			}
			updateTagSlugNames := make([]string, 0)
			for _, tag := range tagList {
				updateTagSlugNames = append(updateTagSlugNames, tag.SlugName)
			}
			err = repo.TagRepo.UpdateTagSynonym(ctx, updateTagSlugNames, converter.StringToInt64(tagInfo.ID), tagInfo.MainTagSlugName)
			if err != nil {
				return err
			}
		}
		revisionDTO.Status = entity.RevisionReviewPassStatus
	} else {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
	}

	tagInfoJson, _ := json.Marshal(tagInfo)
	revisionDTO.Content = string(tagInfoJson)
	revisionID, err := RevisionComServicer.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return err
	}
	if canUpdate {
		ActivityQueueServicer.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         tagInfo.ID,
			OriginalObjectID: tagInfo.ID,
			ActivityTypeKey:  constant.ActTagEdited,
			RevisionID:       revisionID,
		})
	}

	return
}
