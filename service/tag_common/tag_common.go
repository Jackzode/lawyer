package tag_common

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/base/validator"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	entity2 "github.com/lawyer/commons/entity"
	repo "github.com/lawyer/initServer/initRepo"
	"github.com/lawyer/initServer/initServices"
	"sort"
	"strings"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type TagCommonRepo interface {
	AddTagList(ctx context.Context, tagList []*entity2.Tag) (err error)
	GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity2.Tag, err error)
	GetTagBySlugName(ctx context.Context, slugName string) (tagInfo *entity2.Tag, exist bool, err error)
	GetTagListByName(ctx context.Context, name string, recommend, reserved bool) (tagList []*entity2.Tag, err error)
	GetTagListByNames(ctx context.Context, names []string) (tagList []*entity2.Tag, err error)
	GetTagByID(ctx context.Context, tagID string, includeDeleted bool) (tag *entity2.Tag, exist bool, err error)
	GetTagPage(ctx context.Context, page, pageSize int, tag *entity2.Tag, queryCond string) (tagList []*entity2.Tag, total int64, err error)
	GetRecommendTagList(ctx context.Context) (tagList []*entity2.Tag, err error)
	GetReservedTagList(ctx context.Context) (tagList []*entity2.Tag, err error)
	UpdateTagsAttribute(ctx context.Context, tags []string, attribute string, value bool) (err error)
	UpdateTagQuestionCount(ctx context.Context, tagID string, questionCount int) (err error)
}

type TagRepo interface {
	RemoveTag(ctx context.Context, tagID string) (err error)
	UpdateTag(ctx context.Context, tag *entity2.Tag) (err error)
	RecoverTag(ctx context.Context, tagID string) (err error)
	MustGetTagByNameOrID(ctx context.Context, tagID, slugName string) (tag *entity2.Tag, exist bool, err error)
	UpdateTagSynonym(ctx context.Context, tagSlugNameList []string, mainTagID int64, mainTagSlugName string) (err error)
	GetTagSynonymCount(ctx context.Context, tagID string) (count int64, err error)
	GetTagList(ctx context.Context, tag *entity2.Tag) (tagList []*entity2.Tag, err error)
}

type TagRelRepo interface {
	AddTagRelList(ctx context.Context, tagList []*entity2.TagRel) (err error)
	RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	HideTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	RemoveTagRelListByIDs(ctx context.Context, ids []int64) (err error)
	EnableTagRelByIDs(ctx context.Context, ids []int64) (err error)
	GetObjectTagRelWithoutStatus(ctx context.Context, objectId, tagID string) (tagRel *entity2.TagRel, exist bool, err error)
	GetObjectTagRelList(ctx context.Context, objectId string) (tagListList []*entity2.TagRel, err error)
	BatchGetObjectTagRelList(ctx context.Context, objectIds []string) (tagListList []*entity2.TagRel, err error)
	CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error)
}

// TagCommonService user service
type TagCommonService struct {
}

// NewTagCommonService new tag service
func NewTagCommonService() *TagCommonService {
	return &TagCommonService{}
}

// SearchTagLike get tag list all
func (ts *TagCommonService) SearchTagLike(ctx context.Context, req *schema.SearchTagLikeReq) (resp []schema.SearchTagLikeResp, err error) {
	tags, err := repo.TagCommonRepo.GetTagListByName(ctx, req.Tag, len(req.Tag) == 0, false)
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
	mainTagMap := make(map[string]*entity2.Tag)
	if len(mainTagId) > 0 {
		mainTagList, err := repo.TagCommonRepo.GetTagListByIDs(ctx, mainTagId)
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

func (ts *TagCommonService) GetSiteWriteRecommendTag(ctx context.Context) (tags []string, err error) {
	tags = make([]string, 0)
	list, err := repo.TagCommonRepo.GetRecommendTagList(ctx)
	if err != nil {
		return tags, err
	}
	for _, item := range list {
		tags = append(tags, item.SlugName)
	}
	return tags, nil
}

func (ts *TagCommonService) SetSiteWriteTag(ctx context.Context, recommendTags, reservedTags []string, userID string) (
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

func (ts *TagCommonService) GetSiteWriteReservedTag(ctx context.Context) (tags []string, err error) {
	tags = make([]string, 0)
	list, err := repo.TagCommonRepo.GetReservedTagList(ctx)
	if err != nil {
		return tags, err
	}
	for _, item := range list {
		tags = append(tags, item.SlugName)
	}
	return tags, nil
}

// SetTagsAttribute
func (ts *TagCommonService) SetTagsAttribute(ctx context.Context, tags []string, attribute string) (err error) {
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
	err = repo.TagCommonRepo.UpdateTagsAttribute(ctx, tagslist, attribute, false)
	if err != nil {
		return err
	}
	err = repo.TagCommonRepo.UpdateTagsAttribute(ctx, tags, attribute, true)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TagCommonService) GetTagListByNames(ctx context.Context, tagNames []string) ([]*entity2.Tag, error) {
	for k, tagname := range tagNames {
		tagNames[k] = strings.ToLower(tagname)
	}
	tagList, err := repo.TagCommonRepo.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return nil, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return tagList, nil
}

func (ts *TagCommonService) ExistRecommend(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	taginfo, err := services.SiteInfoCommonService.GetSiteWrite(ctx)
	if err != nil {
		return false, err
	}
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

func (ts *TagCommonService) HasNewTag(ctx context.Context, tags []*schema.TagItem) (bool, error) {
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
func (ts *TagCommonService) GetObjectTag(ctx context.Context, objectId string) (objTags []*schema.TagResp, err error) {
	tagsInfoList, err := ts.GetObjectEntityTag(ctx, objectId)
	if err != nil {
		return nil, err
	}
	return ts.TagFormat(ctx, tagsInfoList)
}

// AddTag get object tag
func (ts *TagCommonService) AddTag(ctx context.Context, req *schema.AddTagReq) (resp *schema.AddTagResp, err error) {
	_, exist, err := ts.GetTagBySlugName(ctx, req.SlugName)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.BadRequest(reason.TagAlreadyExist)
	}
	SlugName := strings.ReplaceAll(req.SlugName, " ", "-")
	SlugName = strings.ToLower(SlugName)
	tagInfo := &entity2.Tag{
		SlugName:     SlugName,
		DisplayName:  req.DisplayName,
		OriginalText: req.OriginalText,
		ParsedText:   req.ParsedText,
		Status:       entity2.TagStatusAvailable,
		UserID:       req.UserID,
	}
	tagList := []*entity2.Tag{tagInfo}
	err = repo.TagCommonRepo.AddTagList(ctx, tagList)
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
	_, err = services.RevisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return nil, err
	}
	return &schema.AddTagResp{SlugName: tagInfo.SlugName}, nil
}

// AddTagList get object tag
func (ts *TagCommonService) AddTagList(ctx context.Context, tagList []*entity2.Tag) (err error) {
	return repo.TagCommonRepo.AddTagList(ctx, tagList)
}

// GetTagByID get object tag
func (ts *TagCommonService) GetTagByID(ctx context.Context, tagID string) (tag *entity2.Tag, exist bool, err error) {
	tag, exist, err = repo.TagCommonRepo.GetTagByID(ctx, tagID, false)
	if !exist {
		return
	}
	ts.tagFormatRecommendAndReserved(ctx, tag)
	return
}

// GetTagBySlugName get object tag
func (ts *TagCommonService) GetTagBySlugName(ctx context.Context, slugName string) (tag *entity2.Tag, exist bool, err error) {
	tag, exist, err = repo.TagCommonRepo.GetTagBySlugName(ctx, slugName)
	if !exist {
		return
	}
	ts.tagFormatRecommendAndReserved(ctx, tag)
	return
}

// GetTagListByIDs get object tag
func (ts *TagCommonService) GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity2.Tag, err error) {
	tagList, err = repo.TagCommonRepo.GetTagListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return
}

// GetTagPage get object tag
func (ts *TagCommonService) GetTagPage(ctx context.Context, page, pageSize int, tag *entity2.Tag, queryCond string) (
	tagList []*entity2.Tag, total int64, err error) {
	tagList, total, err = repo.TagCommonRepo.GetTagPage(ctx, page, pageSize, tag, queryCond)
	if err != nil {
		return nil, 0, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return
}

func (ts *TagCommonService) GetObjectEntityTag(ctx context.Context, objectId string) (objTags []*entity2.Tag, err error) {
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

func (ts *TagCommonService) TagFormat(ctx context.Context, tags []*entity2.Tag) (objTags []*schema.TagResp, err error) {
	objTags = make([]*schema.TagResp, 0)
	for _, tagInfo := range tags {
		objTags = append(objTags, &schema.TagResp{
			SlugName:        tagInfo.SlugName,
			DisplayName:     tagInfo.DisplayName,
			MainTagSlugName: tagInfo.MainTagSlugName,
			Recommend:       tagInfo.Recommend,
			Reserved:        tagInfo.Reserved,
		})
	}
	return objTags, nil
}

func (ts *TagCommonService) TagsFormatRecommendAndReserved(ctx context.Context, tagList []*entity2.Tag) {
	if len(tagList) == 0 {
		return
	}
	tagConfig, err := services.SiteInfoCommonService.GetSiteWrite(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	if !tagConfig.RequiredTag {
		for _, tag := range tagList {
			tag.Recommend = false
		}
	}
}

func (ts *TagCommonService) tagFormatRecommendAndReserved(ctx context.Context, tag *entity2.Tag) {
	if tag == nil {
		return
	}
	tagConfig, err := services.SiteInfoCommonService.GetSiteWrite(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	if !tagConfig.RequiredTag {
		tag.Recommend = false
	}
}

// BatchGetObjectTag batch get object tag
func (ts *TagCommonService) BatchGetObjectTag(ctx context.Context, objectIds []string) (map[string][]*schema.TagResp, error) {
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
	tagsInfoMapping := make(map[string]*entity2.Tag)
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

func (ts *TagCommonService) CheckTag(ctx context.Context, tags []string, userID string) (err error) {
	if len(tags) == 0 {
		return nil
	}

	// find tags name
	tagListInDb, err := ts.GetTagListByNames(ctx, tags)
	if err != nil {
		return err
	}

	tagInDbMapping := make(map[string]*entity2.Tag)
	checktags := make([]string, 0)

	for _, tag := range tagListInDb {
		if tag.MainTagID != 0 {
			checktags = append(checktags, fmt.Sprintf("\"%s\"", tag.SlugName))
		}
		tagInDbMapping[tag.SlugName] = tag
	}
	if len(checktags) > 0 {
		err = errors.BadRequest(reason.TagNotContainSynonym).WithMsg(fmt.Sprintf("Should not contain synonym tags %s", strings.Join(checktags, ",")))
		return err
	}

	addTagList := make([]*entity2.Tag, 0)
	addTagMsgList := make([]string, 0)
	for _, tag := range tags {
		_, ok := tagInDbMapping[tag]
		if ok {
			continue
		}
		item := &entity2.Tag{}
		item.SlugName = tag
		item.DisplayName = tag
		item.OriginalText = ""
		item.ParsedText = ""
		item.Status = entity2.TagStatusAvailable
		item.UserID = userID
		addTagList = append(addTagList, item)
		addTagMsgList = append(addTagMsgList, tag)
	}

	if len(addTagList) > 0 {
		err = errors.BadRequest(reason.TagNotFound).WithMsg(fmt.Sprintf("tag [%s] does not exist",
			strings.Join(addTagMsgList, ",")))
		return err

	}

	return nil
}

// CheckTagsIsChange
func (ts *TagCommonService) CheckTagsIsChange(ctx context.Context, tagNameList, oldtagNameList []string) bool {
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

func (ts *TagCommonService) CheckChangeReservedTag(ctx context.Context, oldobjectTagData, objectTagData []*entity2.Tag) (bool, bool, []string, []string) {
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
func (ts *TagCommonService) ObjectChangeTag(ctx context.Context, objectTagData *schema.TagChange) (err error) {
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
	tagListInDb, err := repo.TagCommonRepo.GetTagListByNames(ctx, thisObjTagNameList)
	if err != nil {
		return err
	}

	tagInDbMapping := make(map[string]*entity2.Tag)
	for _, tag := range tagListInDb {
		tagInDbMapping[strings.ToLower(tag.SlugName)] = tag
		thisObjTagIDList = append(thisObjTagIDList, tag.ID)
	}

	addTagList := make([]*entity2.Tag, 0)
	for _, tag := range objectTagData.Tags {
		_, ok := tagInDbMapping[strings.ToLower(tag.SlugName)]
		if ok {
			continue
		}
		item := &entity2.Tag{}
		item.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		item.DisplayName = tag.DisplayName
		item.OriginalText = tag.OriginalText
		item.ParsedText = tag.ParsedText
		item.Status = entity2.TagStatusAvailable
		item.UserID = objectTagData.UserID
		addTagList = append(addTagList, item)
	}

	if len(addTagList) > 0 {
		err = repo.TagCommonRepo.AddTagList(ctx, addTagList)
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
			revisionID, err := services.RevisionService.AddRevision(ctx, revisionDTO, true)
			if err != nil {
				return err
			}
			services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
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

func (ts *TagCommonService) CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error) {
	return repo.TagRelRepo.CountTagRelByTagID(ctx, tagID)
}

// RefreshTagQuestionCount refresh tag question count
func (ts *TagCommonService) RefreshTagQuestionCount(ctx context.Context, tagIDs []string) (err error) {
	for _, tagID := range tagIDs {
		count, err := repo.TagRelRepo.CountTagRelByTagID(ctx, tagID)
		if err != nil {
			return err
		}
		err = repo.TagCommonRepo.UpdateTagQuestionCount(ctx, tagID, int(count))
		if err != nil {
			return err
		}
		log.Debugf("tag count updated %s %d", tagID, count)
	}
	return nil
}

func (ts *TagCommonService) RefreshTagCountByQuestionID(ctx context.Context, questionID string) (err error) {
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
func (ts *TagCommonService) RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.RemoveTagRelListByObjectID(ctx, objectID)
}

// RecoverTagRelListByObjectID recover tag relation by object id
func (ts *TagCommonService) RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.RecoverTagRelListByObjectID(ctx, objectID)
}

func (ts *TagCommonService) HideTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.HideTagRelListByObjectID(ctx, objectID)
}

func (ts *TagCommonService) ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return repo.TagRelRepo.ShowTagRelListByObjectID(ctx, objectID)
}

// CreateOrUpdateTagRelList if tag relation is exists update status, if not create it
func (ts *TagCommonService) CreateOrUpdateTagRelList(ctx context.Context, objectId string, tagIDs []string) (err error) {
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

	addTagRelList := make([]*entity2.TagRel, 0)
	enableTagRelList := make([]int64, 0)
	for _, tagID := range tagIDs {
		needRefreshTagIDs = append(needRefreshTagIDs, tagID)
		rel, exist, err := repo.TagRelRepo.GetObjectTagRelWithoutStatus(ctx, objectId, tagID)
		if err != nil {
			return err
		}
		// if not exist add tag relation
		if !exist {
			addTagRelList = append(addTagRelList, &entity2.TagRel{
				TagID: tagID, ObjectID: objectId, Status: entity2.TagStatusAvailable,
			})
		}
		// if exist and has been removed, that should be enabled
		if exist && rel.Status != entity2.TagStatusAvailable {
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
		log.Error(err)
	}
	return nil
}

func (ts *TagCommonService) UpdateTag(ctx context.Context, req *schema.UpdateTagReq) (err error) {
	var canUpdate bool
	_, existUnreviewed, err := services.RevisionService.ExistUnreviewedByObjectID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if existUnreviewed {
		err = errors.BadRequest(reason.AnswerCannotUpdate)
		return err
	}

	tagInfo, exist, err := ts.GetTagByID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.TagNotFound)
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
			log.Debugf("tag %s update slug_name", tagInfo.SlugName)
			tagList, err := repo.TagRepo.GetTagList(ctx, &entity2.Tag{MainTagID: converter.StringToInt64(tagInfo.ID)})
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
		revisionDTO.Status = entity2.RevisionReviewPassStatus
	} else {
		revisionDTO.Status = entity2.RevisionUnreviewedStatus
	}

	tagInfoJson, _ := json.Marshal(tagInfo)
	revisionDTO.Content = string(tagInfoJson)
	revisionID, err := services.RevisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return err
	}
	if canUpdate {
		services.ActivityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         tagInfo.ID,
			OriginalObjectID: tagInfo.ID,
			ActivityTypeKey:  constant.ActTagEdited,
			RevisionID:       revisionID,
		})
	}

	return
}