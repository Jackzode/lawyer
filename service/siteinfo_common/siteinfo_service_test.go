package siteinfo_common

import (
	"context"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/entity"
	"github.com/lawyer/service/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	mockSiteInfoRepo *mock.MockSiteInfoRepo
)

func mockInit(ctl *gomock.Controller) {
	mockSiteInfoRepo = mock.NewMockSiteInfoRepo(ctl)
	mockSiteInfoRepo.EXPECT().GetByType(gomock.Any(), constant.SiteTypeGeneral).
		Return(&entity.SiteInfo{Content: `{"name":"name"}`}, true, nil)
}

func TestSiteInfoCommonService_GetSiteGeneral(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mockInit(ctl)
	siteInfoCommonService := NewSiteInfoCommonService(mockSiteInfoRepo)
	resp, err := siteInfoCommonService.GetSiteGeneral(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, resp.Name, "name")
}
