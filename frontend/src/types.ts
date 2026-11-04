export type Project = {
  id: number;
  name: string;
  description: string;
  owner_id: number;
};

export type TaskStatus =
  | 'backlog'
  | 'todo'
  | 'in_progress'
  | 'review'
  | 'done';

export type TaskPriority =
  | 'low'
  | 'medium'
  | 'high'
  | 'critical';

export type Task = {
  id: number;
  project_id: number;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee_id: number | null;
  due_date: string | null;
  attachment_path?: string;
};

export const statuses: { key: TaskStatus; label: string }[] = [
  { key: 'backlog', label: 'Бэклог' },
  { key: 'todo', label: 'К выполнению' },
  { key: 'in_progress', label: 'В работе' },
  { key: 'review', label: 'На проверке' },
  { key: 'done', label: 'Готово' },
];

export const priorities: { key: TaskPriority; label: string }[] = [
  { key: 'low', label: 'Низкий' },
  { key: 'medium', label: 'Средний' },
  { key: 'high', label: 'Высокий' },
  { key: 'critical', label: 'Критический' },
];