package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawyer/commons/constant"
	"github.com/lawyer/commons/handler"
	"github.com/lawyer/commons/utils"
	"github.com/lawyer/initServer/initServices"
	"github.com/lawyer/pkg/converter"
	"github.com/lawyer/repo"
	"github.com/lawyer/service/export"
	"io"
	"net/http"
	"net/url"
	"time"
	"xorm.io/xorm/schemas"

	"github.com/lawyer/commons/schema"
	"github.com/lawyer/pkg/dir"
	"github.com/segmentfault/pacman/log"
)

type dashboardService struct {
}

func NewDashboardService() DashboardService {
	return &dashboardService{}
}

type DashboardService interface {
	Statistical(ctx context.Context) (resp *schema.DashboardInfo, err error)
}

func (ds *dashboardService) Statistical(ctx context.Context) (*schema.DashboardInfo, error) {
	dashboardInfo := ds.getFromCache(ctx)
	if dashboardInfo == nil {
		dashboardInfo = &schema.DashboardInfo{}
		dashboardInfo.QuestionCount = ds.questionCount(ctx)
		dashboardInfo.AnswerCount = ds.answerCount(ctx)
		dashboardInfo.CommentCount = ds.commentCount(ctx)
		dashboardInfo.UserCount = ds.userCount(ctx)
		dashboardInfo.ReportCount = ds.reportCount(ctx)
		dashboardInfo.VoteCount = ds.voteCount(ctx)
		dashboardInfo.OccupyingStorageSpace = ds.calculateStorage()
		dashboardInfo.VersionInfo.RemoteVersion = ds.remoteVersion(ctx)
		dashboardInfo.DatabaseVersion = ds.getDatabaseInfo()
		dashboardInfo.DatabaseSize = ds.GetDatabaseSize()
	}

	dashboardInfo.SMTP = ds.smtpStatus(ctx)
	dashboardInfo.HTTPS = ds.httpsStatus(ctx)
	dashboardInfo.TimeZone = ds.getTimezone(ctx)
	dashboardInfo.UploadingFiles = true
	dashboardInfo.AppStartTime = fmt.Sprintf("%d", time.Now().Unix()-schema.AppStartTime.Unix())
	dashboardInfo.VersionInfo.Version = constant.Version
	dashboardInfo.VersionInfo.Revision = constant.Revision
	dashboardInfo.GoVersion = constant.GoVersion
	if siteLogin, err := services.SiteInfoService.GetSiteLogin(ctx); err == nil {
		dashboardInfo.LoginRequired = siteLogin.LoginRequired
	}

	ds.setCache(ctx, dashboardInfo)
	return dashboardInfo, nil
}

func (ds *dashboardService) getFromCache(ctx context.Context) (dashboardInfo *schema.DashboardInfo) {
	infoStr := handler.RedisClient.Get(ctx, schema.DashboardCacheKey).String()
	if infoStr == "" {
		return nil
	}
	dashboardInfo = &schema.DashboardInfo{}
	if err := json.Unmarshal([]byte(infoStr), dashboardInfo); err != nil {
		return nil
	}
	return dashboardInfo
}

func (ds *dashboardService) setCache(ctx context.Context, info *schema.DashboardInfo) {
	infoStr, _ := json.Marshal(info)
	err := handler.RedisClient.Set(ctx, schema.DashboardCacheKey, string(infoStr), schema.DashboardCacheTime).Err()
	if err != nil {
		log.Errorf("set dashboard statistical failed: %s", err)
	}
}

func (ds *dashboardService) questionCount(ctx context.Context) int64 {
	questionCount, err := repo.QuestionRepo.GetQuestionCount(ctx)
	if err != nil {
		log.Errorf("get question count failed: %s", err)
	}
	return questionCount
}

func (ds *dashboardService) answerCount(ctx context.Context) int64 {
	answerCount, err := repo.AnswerRepo.GetAnswerCount(ctx)
	if err != nil {
		log.Errorf("get answer count failed: %s", err)
	}
	return answerCount
}

func (ds *dashboardService) commentCount(ctx context.Context) int64 {
	commentCount, err := repo.CommentCommonRepo.GetCommentCount(ctx)
	if err != nil {
		log.Errorf("get comment count failed: %s", err)
	}
	return commentCount
}

func (ds *dashboardService) userCount(ctx context.Context) int64 {
	userCount, err := repo.UserRepo.GetUserCount(ctx)
	if err != nil {
		log.Errorf("get user count failed: %s", err)
	}
	return userCount
}

func (ds *dashboardService) reportCount(ctx context.Context) int64 {
	reportCount, err := repo.ReportRepo.GetReportCount(ctx)
	if err != nil {
		log.Errorf("get report count failed: %s", err)
	}
	return reportCount
}

// count vote
func (ds *dashboardService) voteCount(ctx context.Context) int64 {
	typeKeys := []string{
		"question.vote_up",
		"question.vote_down",
		"answer.vote_up",
		"answer.vote_down",
	}
	var activityTypes []int
	for _, typeKey := range typeKeys {
		cfg, err := utils.GetConfigByKey(ctx, typeKey)
		if err != nil {
			continue
		}
		activityTypes = append(activityTypes, cfg.ID)
	}
	voteCount, err := repo.VoteRepo.GetVoteCount(ctx, activityTypes)
	if err != nil {
		log.Errorf("get vote count failed: %s", err)
	}
	return voteCount
}

func (ds *dashboardService) remoteVersion(ctx context.Context) string {
	req, _ := http.NewRequest("GET", "https://getlatest.answer.dev/", nil)
	req.Header.Set("User-Agent", "Answer/"+constant.Version)
	httpClient := &http.Client{}
	httpClient.Timeout = 15 * time.Second
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("request remote version failed: %s", err)
		return ""
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("read response body failed: %s", err)
		return ""
	}
	remoteVersion := &schema.RemoteVersion{}
	if err := json.Unmarshal(respByte, remoteVersion); err != nil {
		log.Errorf("parsing response body failed: %s", err)
		return ""
	}
	return remoteVersion.Release.Version
}

func (ds *dashboardService) smtpStatus(ctx context.Context) (smtpStatus string) {
	emailConf, err := utils.GetStringValue(ctx, "email.config")
	if err != nil {
		log.Errorf("get email config failed: %s", err)
		return "disabled"
	}
	ec := &export.EmailConfig{}
	err = json.Unmarshal([]byte(emailConf), ec)
	if err != nil {
		log.Errorf("parsing email config failed: %s", err)
		return "disabled"
	}
	if ec.SMTPHost != "" {
		smtpStatus = "enabled"
	}
	return smtpStatus
}

func (ds *dashboardService) httpsStatus(ctx context.Context) (enabled bool) {
	siteGeneral, err := services.SiteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site general failed: %s", err)
		return false
	}
	siteUrl, err := url.Parse(siteGeneral.SiteUrl)
	if err != nil {
		log.Errorf("parse site url failed: %s", err)
		return false
	}
	return siteUrl.Scheme == "https"
}

func (ds *dashboardService) getTimezone(ctx context.Context) string {
	siteInfoInterface, err := services.SiteInfoService.GetSiteInterface(ctx)
	if err != nil {
		return ""
	}
	return siteInfoInterface.TimeZone
}

func (ds *dashboardService) calculateStorage() string {
	dirSize, err := dir.DirSize("ds.serviceConfig.UploadPath")
	if err != nil {
		log.Errorf("get upload dir size failed: %s", err)
		return ""
	}
	return dir.FormatFileSize(dirSize)
}

func (ds *dashboardService) getDatabaseInfo() (versionDesc string) {
	dbVersion, err := handler.Engine.DBVersion()
	if err != nil {
		log.Errorf("get db version failed: %s", err)
	} else {
		versionDesc = fmt.Sprintf("%s %s", handler.Engine.Dialect().URI().DBType, dbVersion.Number)
	}
	return versionDesc
}

func (ds *dashboardService) GetDatabaseSize() (dbSize string) {
	switch handler.Engine.Dialect().URI().DBType {
	case schemas.MYSQL:
		sql := fmt.Sprintf("SELECT SUM(DATA_LENGTH) as db_size FROM information_schema.TABLES WHERE table_schema = '%s'",
			handler.Engine.Dialect().URI().DBName)
		res, err := handler.Engine.QueryInterface(sql)
		if err != nil {
			log.Warnf("get db size failed: %s", err)
		} else {
			if res != nil && len(res) > 0 && res[0]["db_size"] != nil {
				dbSizeStr, _ := res[0]["db_size"].(string)
				dbSize = dir.FormatFileSize(converter.StringToInt64(dbSizeStr))
			}
		}
	case schemas.POSTGRES:
		sql := fmt.Sprintf("SELECT pg_database_size('%s') AS db_size",
			handler.Engine.Dialect().URI().DBName)
		res, err := handler.Engine.QueryInterface(sql)
		if err != nil {
			log.Warnf("get db size failed: %s", err)
		} else {
			if res != nil && len(res) > 0 && res[0]["db_size"] != nil {
				dbSizeStr, _ := res[0]["db_size"].(int32)
				dbSize = dir.FormatFileSize(int64(dbSizeStr))
			}
		}
	case schemas.SQLITE:
		dirSize, err := dir.DirSize(handler.Engine.DataSourceName())
		if err != nil {
			log.Errorf("get upload dir size failed: %s", err)
			return ""
		}
		dbSize = dir.FormatFileSize(dirSize)
	}
	return dbSize
}
