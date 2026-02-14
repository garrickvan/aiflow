package api

import (
	"aiflow/internal/api/handlers"
	"aiflow/internal/repositories"
	"aiflow/internal/services"

	"github.com/go-chi/chi/v5"
)

// Router API路由器
type Router struct {
	skillHandler   *handlers.SkillHandler
	tagHandler     *handlers.TagHandler
	uploadHandler  *handlers.UploadHandler
	jobTaskHandler *handlers.JobTaskHandler
}

// NewRouter 创建新的API路由器
func NewRouter(repo *repositories.Repository) *Router {
	// 初始化service层
	skillService := services.NewSkillService(repo)
	tagService := services.NewTagService(repo)
	jobTaskService := services.NewJobTaskService(repo)

	return &Router{
		skillHandler:   handlers.NewSkillHandler(skillService),
		tagHandler:     handlers.NewTagHandler(tagService),
		uploadHandler:  handlers.NewUploadHandler(repo),
		jobTaskHandler: handlers.NewJobTaskHandler(jobTaskService),
	}
}

// RegisterRoutes 注册API路由
func (r *Router) RegisterRoutes(chiRouter chi.Router) {
	// API根路径
	chiRouter.Route("/api", func(api chi.Router) {
		// 标签相关路由
		api.Route("/tags", func(tags chi.Router) {
			tags.Get("/", r.tagHandler.ListTags)         // 获取所有标签
			tags.Post("/", r.tagHandler.CreateTag)       // 创建标签
			tags.Get("/{id}", r.tagHandler.GetTag)       // 根据ID获取标签
			tags.Put("/{id}", r.tagHandler.UpdateTag)    // 更新标签
			tags.Delete("/{id}", r.tagHandler.DeleteTag) // 删除标签
		})

		// 技能相关路由
		api.Route("/skills", func(skills chi.Router) {
			skills.Get("/", r.skillHandler.ListSkills)              // 获取所有技能
			skills.Post("/", r.skillHandler.CreateSkill)            // 创建技能
			skills.Get("/trash", r.skillHandler.ListDeletedSkills)  // 获取回收站技能列表
			skills.Get("/{id}", r.skillHandler.GetSkill)            // 根据ID获取技能
			skills.Put("/{id}", r.skillHandler.UpdateSkill)         // 更新技能
			skills.Delete("/{id}", r.skillHandler.DeleteSkill)      // 删除技能（伪删除，进入回收站）
			skills.Post("/{id}/restore", r.skillHandler.RestoreSkill) // 恢复回收站中的技能
			skills.Delete("/{id}/permanent", r.skillHandler.PermanentDeleteSkill) // 彻底删除技能
			skills.Get("/export", r.skillHandler.ExportSkills)      // 导出所有技能为MD格式
			skills.Get("/{id}/export", r.skillHandler.ExportSkills) // 导出单个技能为MD格式
		})

		// 文件上传路由
		api.Post("/upload_data", r.uploadHandler.UploadData) // 上传文件

		// 任务相关路由
		api.Route("/jobtasks", func(jobtasks chi.Router) {
			jobtasks.Get("/", r.jobTaskHandler.ListJobTasks)                     // 获取任务列表
			jobtasks.Post("/", r.jobTaskHandler.CreateJobTask)                   // 创建任务
			jobtasks.Post("/export", r.jobTaskHandler.BatchExportJobTasks)       // 批量导出任务
			jobtasks.Get("/projects", r.jobTaskHandler.GetAllJobTaskProjects)    // 获取所有项目列表（去重）
			jobtasks.Get("/trash", r.jobTaskHandler.ListDeletedJobTasks)         // 获取回收站列表
			jobtasks.Get("/{id}", r.jobTaskHandler.GetJobTask)                   // 根据ID获取任务
			jobtasks.Put("/{id}", r.jobTaskHandler.UpdateJobTask)                // 更新任务
			jobtasks.Delete("/{id}", r.jobTaskHandler.DeleteJobTask)             // 删除任务（伪删除，进入回收站）
			jobtasks.Post("/{id}/restore", r.jobTaskHandler.RestoreJobTask)      // 恢复回收站中的任务
			jobtasks.Delete("/{id}/permanent", r.jobTaskHandler.PermanentDeleteJobTask) // 彻底删除任务
		})
	})
}
