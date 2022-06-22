package job

import (
	"context"
)

// 本地任务
func Start(ctx context.Context) error {
	m := metadata.FromContext(ctx)
	sessionId := m.GetReqHeaderParam().SessionId
	if sessionId == "" {
		return ecode.SessionIsEmpty
	}
	s, ok := model.GetSessionById(sessionId)
	if !ok {
		return ecode.SessionNotExist
	}
	// 获取 session 中任务
	tasks := model.GetSessionTasks(sessionId)
	if len(tasks) == 0 {
		return ecode.TaskIsEmpty
	}

	// 检查当前是否允许下单
	if t := dispatcher.CurrTask(); t != nil && t.GetTaskType() != worker.CallTaskType {
		return ecode.TaskIsBusy
	}

	// 获取配置
	c := dispatcher.StartqConfig{}
	c.SessionId = sessionId

	siteConfig := model.GetSiteConfig(ctx)
	c.HomeId = siteConfig.HomeId
	c.Speed = siteConfig.HomeSpeed

tasksLoop:
	for _, task := range tasks {
		// 递送单不使用jobid作为任务id
		if s.AppName != model.DeliveryAppName {
			if job.FindTask(task.TaskId) != -1 {
				continue
			}
		}

		if s.AppName == model.DeliveryAppName {
			d, ok := model.GetDeliveryById(task.TaskId)
			if !ok {
				model.DeleteSessionTask(sessionId, task.TaskId)
				continue
			}

			// 获取所有任务相关的配送单
			deliverys := make([]model.Delivery, 0)
			taskIds := job.GetTaskIdsByType(worker.DeliveryTaskType)
			for _, taskId := range taskIds {
				delis := model.GetDeliveryByTaskId(taskId)
				for _, d := range delis {
					deliverys = append(deliverys, d)
				}
			}

			for _, item := range deliverys {
				if model.IsSameDelivery(item, d) {
					model.SetDeliveryTaskId(d.DeliveryId, item.TaskId)
					continue tasksLoop
				}
			}

			taskId := utils.GenSnowFlakeIdStr(runtime.GetRobotConfig().RobotID)
			t := worker.NewDeliveryTask(sessionId, taskId, task.PoiId, task.CreatedAt)
			model.SetDeliveryTaskId(task.TaskId, taskId)
			job.Putq(t)
		}

		if s.AppName == model.LeadwayAppName {
			t := worker.NewLeadwayTask(sessionId, task.TaskId, task.PoiId, task.CreatedAt, worker.LeadwayPriority)
			job.Putq(t)
		}

		if s.AppName == model.CruiseAppName {
			t := worker.NewCruiseTask(sessionId, task.TaskId)
			job.Putq(t)
			// 清除任务记录
			model.DeleteSessionTask(sessionId, task.TaskId)
		}

		if s.AppName == model.DisinfectionAppName {
			t := worker.NewDisinfectionTask(sessionId, task.TaskId)
			job.Putq(t)
			// 清除任务记录
			model.DeleteSessionTask(sessionId, task.TaskId)
		}
	}

	dispatcher.Startq(c)
	return nil
}

func Clear(ctx context.Context) error {
	job.Clear()
	return nil
}

func Cancel(ctx context.Context) error {
	if dispatcher.CurrTask() == nil {
		return ecode.TaskIsEmpty
	}
	if dispatcher.CurrTask().Cancel() != nil {
		return ecode.CancelTaskFailed
	}
	return nil
}

func Pause(ctx context.Context) error {
	if dispatcher.CurrTask() != nil {
		if dispatcher.CurrTask().Pause() != nil {
			return ecode.PauseTaskFailed
		}
	}

	return nil
}

func Resume(ctx context.Context) error {
	if dispatcher.CurrTask() != nil {
		return dispatcher.CurrTask().Resume()
	}
	return nil
}

func StartNext(ctx context.Context) error {
	if dispatcher.CurrTask() == nil {
		return ecode.TaskIsEmpty
	}
	dispatcher.CurrTask().StartNext()
	return nil
}

type AppraiseRequest struct {
	JobIds []string `json:"job_ids"`
	Score  int32    `json:"score"`
}

func Appraise(ctx context.Context, in AppraiseRequest) error {
	m := metadata.FromContext(ctx)
	sessionId := m.GetReqHeaderParam().SessionId
	if sessionId == "" {
		return ecode.SessionIsEmpty
	}
	s, ok := model.GetSessionById(sessionId)
	if !ok {
		return ecode.SessionNotExist
	}

	// 递送
	if s.AppName == model.DeliveryAppName {
		for _, deliveryId := range in.JobIds {
			if !model.SetDeliveryScore(deliveryId, in.Score) {
				bsm.UpdateJobScore(ctx, deliveryId, in.Score)
			}
		}
	}

	// 引领
	if s.AppName == model.LeadwayAppName {
		for _, leadwayId := range in.JobIds {
			if !model.SetLeadwayScore(leadwayId, in.Score) {
				bsm.UpdateJobScore(ctx, leadwayId, in.Score)
			}
		}
	}

	return nil
}
