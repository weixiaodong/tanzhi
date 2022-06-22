// 定时处理检测任务
package crontab

import (
	"context"
	"encoding/json"

	"github.com/robfig/cron/v3"

	"github.com/weixiaodong/tanzhi/component/log"
	"github.com/weixiaodong/tanzhi/config"
	"github.com/weixiaodong/tanzhi/crontab/handler"
	"github.com/weixiaodong/tanzhi/internal/notifych"
	"github.com/weixiaodong/tanzhi/internal/store/db"
)

func Start() {
	ctx := context.Background()

	c := cron.New(
		cron.WithSeconds(),
		cron.WithChain(
			cron.Recover(cron.DefaultLogger),
		),
	)
	var err error

	// 启动配置文件中配置的任务
	for _, jobCfg := range config.GConfig.Jobs {
		_, err = c.AddJob(
			jobCfg.Expr,
			handler.NewCornJob(
				ctx,
				jobCfg.Name,
				jobCfg.Command.Type,
				jobCfg.Command.Method,
				jobCfg.Command.Target,
			),
		)

		if err != nil {
			log.Error(ctx, "add_job_failed", "load config job error ", err, "jobCfg", jobCfg)
			continue
		}
		log.Info(ctx, "add_job_successed", "jobCfg", jobCfg)
	}

	// 启动db中配置的任务
	dbJobs, err := db.GetJob(ctx)
	if err != nil {
		log.Error(ctx, "add_job_failed", "load db job error ", err)
	} else {
		for _, dbJob := range dbJobs {
			command := &config.CommandConfig{}
			err := json.Unmarshal([]byte(dbJob.Command), command)
			if err != nil {
				log.Error(ctx, "add_job_failed", "load db job error ", err, "dbJob", dbJob)
				continue
			}
			_, err = c.AddJob(
				dbJob.Expr,
				handler.NewCornJob(
					ctx,
					dbJob.Name,
					command.Type,
					command.Method,
					command.Target,
				),
			)

			if err != nil {
				log.Error(ctx, "add_db_job_failed", "load config job error ", err, "dbJob", dbJob)
				continue
			}
			log.Info(ctx, "add_db_job_successed", "dbJob", dbJob)
		}
	}

	// 异步启动http接口中创建的任务
	go func() {
		// 根据notifych中的JobCreateCh获取创建任务
		for jobCfg := range notifych.JobCreateCh {
			_, err = c.AddJob(
				jobCfg.Expr,
				handler.NewCornJob(
					ctx,
					jobCfg.Name,
					jobCfg.Command.Type,
					jobCfg.Command.Method,
					jobCfg.Command.Target,
				),
			)
			if err != nil {
				log.Error(ctx, "add_httpcreate_job_failed", "load config job error ", err, "jobCfg", jobCfg)
				continue
			}
			log.Info(ctx, "add_httpcreate_job_successed", "jobCfg", jobCfg)
		}
	}()

	c.Start()
}
