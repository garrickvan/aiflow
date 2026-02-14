/**
 * API服务层
 * 用于对接后端API接口
 */

import { API_BASE_URL } from '../types';
import type {
  Tag,
  Skill,
  SkillRequest,
  JobTask,
  JobTaskRequest,
  PaginatedResponse,
} from '../types';

// 重新导出类型，供其他模块使用
export type { Tag, Skill, SkillRequest, JobTask, JobTaskRequest, PaginatedResponse };

/**
 * 通用请求函数
 * @param url - 请求路径
 * @param options - 请求选项
 * @returns 响应数据
 */
async function request<T>(url: string, options?: RequestInit): Promise<T> {
  try {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    if (!data.success) {
      throw new Error(data.error || '操作失败');
    }

    return data.data as T;
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
}

/**
 * 标签API
 */
export const tagApi = {
  /**
   * 获取所有标签（支持分页）
   */
  async getTags(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<Tag>> {
    return request<PaginatedResponse<Tag>>(`/tags?page=${page}&pageSize=${pageSize}`);
  },

  /**
   * 创建标签
   */
  async createTag(name: string): Promise<Tag> {
    return request<Tag>('/tags', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  },

  /**
   * 更新标签
   */
  async updateTag(id: number, name: string): Promise<Tag> {
    return request<Tag>(`/tags/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    });
  },

  /**
   * 删除标签
   */
  async deleteTag(id: number): Promise<void> {
    await request<void>(`/tags/${id}`, {
      method: 'DELETE',
    });
  },
};

/**
 * 技能API
 */
export const skillApi = {
  /**
   * 获取所有技能（支持分页、标签筛选和日期范围筛选）
   */
  async getSkills(
    tagId?: number,
    page: number = 1,
    pageSize: number = 10,
    startDate?: number,
    endDate?: number
  ): Promise<PaginatedResponse<Skill>> {
    let url = '/skills';
    const params = new URLSearchParams();
    if (tagId) {
      params.append('tagId', tagId.toString());
    }
    if (startDate) {
      params.append('startDate', startDate.toString());
    }
    if (endDate) {
      params.append('endDate', endDate.toString());
    }
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    const queryString = params.toString();
    if (queryString) {
      url += `?${queryString}`;
    }

    return request<PaginatedResponse<Skill>>(url);
  },

  /**
   * 创建技能
   */
  async createSkill(skill: SkillRequest): Promise<Skill> {
    return request<Skill>('/skills', {
      method: 'POST',
      body: JSON.stringify(skill),
    });
  },

  /**
   * 更新技能
   */
  async updateSkill(id: number, skill: SkillRequest): Promise<Skill> {
    return request<Skill>(`/skills/${id}`, {
      method: 'PUT',
      body: JSON.stringify(skill),
    });
  },

  /**
   * 删除技能（伪删除，进入回收站）
   */
  async deleteSkill(id: number): Promise<void> {
    await request<void>(`/skills/${id}`, {
      method: 'DELETE',
    });
  },

  /**
   * 获取回收站技能列表
   */
  async getTrashSkills(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<Skill>> {
    let url = '/skills/trash';
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());
    return request<PaginatedResponse<Skill>>(`${url}?${params.toString()}`);
  },

  /**
   * 恢复回收站中的技能
   */
  async restoreSkill(id: number): Promise<void> {
    await request<void>(`/skills/${id}/restore`, {
      method: 'POST',
    });
  },

  /**
   * 彻底删除技能（从回收站中永久删除）
   */
  async permanentDeleteSkill(id: number): Promise<void> {
    await request<void>(`/skills/${id}/permanent`, {
      method: 'DELETE',
    });
  },

  /**
   * 导出技能
   */
  async exportSkill(id: number): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/skills/${id}/export`);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const filename = 'SKILL.md';

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export failed:', error);
      throw error;
    }
  },
};

/**
 * 任务API
 */
export const jobtaskApi = {
  /**
   * 获取任务列表（支持分页、多条件筛选和日期范围筛选）
   */
  async getJobTasks(
    project?: string,
    jobType?: string,
    status?: string,
    page: number = 1,
    pageSize: number = 10,
    startDate?: number,
    endDate?: number
  ): Promise<PaginatedResponse<JobTask>> {
    let url = '/jobtasks';
    const params = new URLSearchParams();
    if (project && project !== '') {
      params.append('project', project);
    }
    if (jobType && jobType !== '') {
      params.append('type', jobType);
    }
    if (status && status !== '') {
      params.append('status', status);
    }
    if (startDate) {
      params.append('startDate', startDate.toString());
    }
    if (endDate) {
      params.append('endDate', endDate.toString());
    }
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    const queryString = params.toString();
    if (queryString) {
      url += `?${queryString}`;
    }

    return request<PaginatedResponse<JobTask>>(url);
  },

  /**
   * 获取所有项目列表（去重）
   */
  async getProjects(): Promise<string[]> {
    return request<string[]>('/jobtasks/projects');
  },

  /**
   * 创建任务
   */
  async createJobTask(jobTask: JobTaskRequest): Promise<JobTask> {
    return request<JobTask>('/jobtasks', {
      method: 'POST',
      body: JSON.stringify(jobTask),
    });
  },

  /**
   * 更新任务
   */
  async updateJobTask(id: number, jobTask: JobTaskRequest): Promise<JobTask> {
    return request<JobTask>(`/jobtasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(jobTask),
    });
  },

  /**
   * 删除任务（伪删除，进入回收站）
   */
  async deleteJobTask(id: number): Promise<void> {
    await request<void>(`/jobtasks/${id}`, {
      method: 'DELETE',
    });
  },

  /**
   * 获取回收站任务列表
   */
  async getTrashJobTasks(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<JobTask>> {
    let url = '/jobtasks/trash';
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());
    return request<PaginatedResponse<JobTask>>(`${url}?${params.toString()}`);
  },

  /**
   * 恢复回收站中的任务
   */
  async restoreJobTask(id: number): Promise<void> {
    await request<void>(`/jobtasks/${id}/restore`, {
      method: 'POST',
    });
  },

  /**
   * 彻底删除任务（从回收站中永久删除）
   */
  async permanentDeleteJobTask(id: number): Promise<void> {
    await request<void>(`/jobtasks/${id}/permanent`, {
      method: 'DELETE',
    });
  },

  /**
   * 批量导出任务
   */
  async exportJobTasks(ids?: number[], format: 'csv' | 'json' | 'md' = 'csv'): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/jobtasks/export`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ ids, format }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
      const extensionMap = { csv: 'csv', json: 'json', md: 'md' } as const;
      const filename = `jobtasks_${timestamp}.${extensionMap[format]}`;

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export failed:', error);
      throw error;
    }
  },
};

/**
 * 文件上传API
 */
export const uploadApi = {
  /**
   * 上传文件
   */
  async uploadFile(file: File, processType: string): Promise<{ success: boolean; message: string }> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('process_type', processType);

    try {
      const response = await fetch(`${API_BASE_URL}/upload_data`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      if (!data.success) {
        throw new Error(data.error || '上传失败');
      }

      return data;
    } catch (error) {
      console.error('Upload failed:', error);
      throw error;
    }
  },
};
