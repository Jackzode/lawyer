package cron

import (
	"context"
	"fmt"

	"github.com/lawyer/service"
	"github.com/robfig/cron/v3"
	"github.com/segmentfault/pacman/log"
)

// ScheduledTaskManager scheduled task manager
type ScheduledTaskManager struct {
	siteInfoService service.SiteInfoCommonServicer
	questionService *service.QuestionServicer
}

// NewScheduledTaskManager new scheduled task manager
func NewScheduledTaskManager(
	siteInfoService service.SiteInfoCommonServicer,
	questionService *service.QuestionServicer,
) *ScheduledTaskManager {
	manager := &ScheduledTaskManager{
		siteInfoService: siteInfoService,
		questionService: questionService,
	}
	return manager
}

func (s *ScheduledTaskManager) Run() {
	fmt.Println("start cron")
	s.questionService.SitemapCron(context.Background())
	c := cron.New()
	_, err := c.AddFunc("0 */1 * * *", func() {
		ctx := context.Background()
		fmt.Println("sitemap cron execution")
		s.questionService.SitemapCron(ctx)
	})
	if err != nil {
		log.Error(err)
	}
	c.Start()
}
