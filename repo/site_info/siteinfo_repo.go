package site_info

import (
	"context"
	"encoding/json"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/constant/reason"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/commons/handler"
	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"

	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
)

type siteInfoRepo struct {
	DB    *xorm.Engine
	Cache *redis.Client
}

func NewSiteInfo() *siteInfoRepo {
	return &siteInfoRepo{
		DB:    handler.Engine,
		Cache: handler.RedisClient,
	}
}

// SaveByType save site setting by type
func (sr *siteInfoRepo) SaveByType(ctx context.Context, siteType string, data *entity.SiteInfo) (err error) {
	old := &entity.SiteInfo{}
	exist, err := sr.DB.Context(ctx).Where(builder.Eq{"type": siteType}).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		_, err = sr.DB.Context(ctx).ID(old.ID).Update(data)
	} else {
		_, err = sr.DB.Context(ctx).Insert(data)
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	sr.setCache(ctx, siteType, data)
	return
}

// GetByType get site info by type
func (sr *siteInfoRepo) GetByType(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo, exist bool, err error) {
	siteInfo = sr.getCache(ctx, siteType)
	if siteInfo != nil {
		return siteInfo, true, nil
	}
	siteInfo = &entity.SiteInfo{}
	exist, err = sr.DB.Context(ctx).Where(builder.Eq{"type": siteType}).Get(siteInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return nil, false, err
	}
	if exist {
		sr.setCache(ctx, siteType, siteInfo)
	}
	return
}

func (sr *siteInfoRepo) getCache(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo) {
	siteInfoCache := sr.Cache.Get(ctx, constant.SiteInfoCacheKey+siteType).String()
	if siteInfoCache == "" {
		return nil
	}
	siteInfo = &entity.SiteInfo{}
	_ = json.Unmarshal([]byte(siteInfoCache), siteInfo)
	return siteInfo
}

func (sr *siteInfoRepo) setCache(ctx context.Context, siteType string, siteInfo *entity.SiteInfo) {
	siteInfoCache, _ := json.Marshal(siteInfo)
	err := sr.Cache.Set(ctx,
		constant.SiteInfoCacheKey+siteType, string(siteInfoCache), constant.SiteInfoCacheTime).Err()
	if err != nil {
		log.Error(err)
	}
}
