package middleware

import (
	"github.com/gin-gonic/gin"
)

type ShortIDMiddleware struct {
	//siteInfoService siteinfo_common.SiteInfoCommonService
}

func NewShortIDMiddleware(
// siteInfoService siteinfo_common.SiteInfoCommonService
) *ShortIDMiddleware {
	return &ShortIDMiddleware{
		//siteInfoService: siteInfoService,
	}
}

func (sm *ShortIDMiddleware) SetShortIDFlag() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//siteSeo, err := sm.siteInfoService.GetSiteSeo(ctx)
		//if err != nil {
		//	log.Error(err)
		//	return
		//}
		//ctx.Set(constant.ShortIDFlag, siteSeo.IsShortLink())
	}
}
