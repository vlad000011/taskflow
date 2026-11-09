import { useEffect, useMemo, useState } from 'react';
import type { FormEvent, ReactNode } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  DndContext,
  type DragEndEvent,
  useDraggable,
  useDroppable,
} from '@dnd-kit/core';

import {
  Activity,
  BarChart3,
  CheckCircle2,
  CircleHelp,
  ClipboardList,
  FolderKanban,
  LogOut,
  Moon,
  Plus,
  Pencil,
  Trash2,
  Search,
  Settings,
  ShieldCheck,
  Sun,
  Users,
  User,
  Upload as UploadIcon,
  Download,
  X,
} from 'lucide-react';

import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import { api } from './lib/api';
import { useAuth, type User as UserType } from './stores/auth';
import {
  statuses,
  type Project,
  type Task,
  type TaskStatus,
} from './types';

const priorities: {
  key: Task['priority'];
  label: string;
}[] = [
  { key: 'low', label: 'Низкий' },
  { key: 'medium', label: 'Средний' },
  { key: 'high', label: 'Высокий' },
  { key: 'critical', label: 'Критический' },
];

function priorityLabel(priority: Task['priority']) {
  return priorities.find(p => p.key === priority)?.label ?? 'Средний';
}

function App() {
  const { user, setUser, logout } = useAuth();

  const [dark, setDark] = useState(
    localStorage.getItem('taskflow_dark') !== 'false'
  );

  const [authLoading, setAuthLoading] = useState(true);
  const [email, setEmail] = useState('lead@taskflow.local');
  const [password, setPassword] = useState('ChangeMe123!');
  const [error, setError] = useState('');

  const [newProject, setNewProject] = useState('');
  const [projectID, setProjectID] = useState<number | null>(null);

  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [assigneeFilter, setAssigneeFilter] = useState('');
  const [sort, setSort] = useState('created_at');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [taskPage, setTaskPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const [showCreate, setShowCreate] = useState(false);
  const [createKind, setCreateKind] = useState<'project' | 'task'>('task');
  const [createError, setCreateError] = useState('');
  const [newTask, setNewTask] = useState('');
  const [newPriority, setNewPriority] =
    useState<Task['priority']>('medium');
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [editError, setEditError] = useState('');
  const [uploadingTaskIds, setUploadingTaskIds] = useState<number[]>([]);

  const [activePage, setActivePage] = useState('board');

  const [activeModal, setActiveModal] =
    useState<'help' | 'settings' | null>(null);

  const [profileOpen, setProfileOpen] = useState(false);
  const [profileName, setProfileName] = useState(user?.name ?? '');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [profileError, setProfileError] = useState('');
  const [profileSuccess, setProfileSuccess] = useState('');

  const qc = useQueryClient();
  const nav = useNavigate();

  useEffect(() => {
    if (user) setProfileName(user.name);
  }, [user?.name]);

  useEffect(() => {
    document.documentElement.dataset.theme = dark ? 'dark' : 'light';
    localStorage.setItem('taskflow_dark', String(dark));
  }, [dark]);

  useEffect(() => {
    const token = localStorage.getItem('taskflow_access');

    if (!token) {
      setAuthLoading(false);
      return;
    }

    api
      .get('/users/me')
      .then(r => setUser(r.data.data as UserType))
      .catch(() => logout())
      .finally(() => setAuthLoading(false));
  }, [setUser, logout]);

  const projectsQuery = useQuery({
    queryKey: ['projects'],
    queryFn: async () =>
      (await api.get('/projects')).data.data as Project[],
    enabled: !!user,
  });

  const projects = projectsQuery.data ?? [];

  const currentProject =
    projects.find(p => p.id === projectID) ?? projects[0];

  useEffect(() => {
    if (currentProject && !projectID) {
      setProjectID(currentProject.id);
    }
  }, [currentProject, projectID]);

  const membersQuery = useQuery({
    queryKey: ['project-members', currentProject?.id],
    queryFn: async () => (await api.get(`/projects/${currentProject!.id}/members`)).data.data as { id: number; name: string; email: string }[],
    enabled: !!user && !!currentProject,
  });

  const tasksQuery = useQuery({
    queryKey: ['tasks', currentProject?.id, search, filter, statusFilter, assigneeFilter, sort, sortOrder, taskPage, pageSize],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: String(taskPage),
        limit: String(pageSize),
        search,
        sort,
        order: sortOrder,
      });
      if (filter) params.set('priority', filter);
      if (statusFilter) params.set('status', statusFilter);
      if (assigneeFilter) params.set('assignee_id', assigneeFilter);

      const result = await api.get(`/projects/${currentProject!.id}/tasks?${params}`);
      return result.data.data as {
        items: Task[];
        pagination: { page: number; limit: number; total: number; pages: number };
      };
    },
    enabled: !!user && !!currentProject,
  });

  const dashboardQuery = useQuery({
    queryKey: ['dashboard', currentProject?.id],
    queryFn: async () =>
      (
        await api.get(
          `/projects/${currentProject!.id}/dashboard`
        )
      ).data.data,
    enabled: !!user && !!currentProject,
  });

  const tasks = tasksQuery.data?.items ?? [];
  const pagination = tasksQuery.data?.pagination;
  const totalPages = Math.max(1, pagination?.pages ?? 1);
  const dashboard = dashboardQuery.data;

  useEffect(() => {
    if (pagination && taskPage > totalPages) setTaskPage(totalPages);
  }, [pagination, taskPage, totalPages]);

  async function saveProfile(e: FormEvent) {
    e.preventDefault();
    if (!profileName.trim()) { setProfileError('Имя не может быть пустым.'); return; }
    setProfileError('');
    setProfileSuccess('');
    try {
      const response = await api.patch('/users/me', { name: profileName.trim() });
      setUser(response.data.data as UserType);
      setProfileSuccess('Имя успешно обновлено.');
    } catch (error: any) {
      setProfileError(error?.response?.data?.error?.message ?? 'Не удалось обновить профиль.');
    }
  }

  async function changeOwnPassword(e: FormEvent) {
    e.preventDefault();
    setProfileError('');
    setProfileSuccess('');
    if (newPassword.length < 8) { setProfileError('Новый пароль должен содержать минимум 8 символов.'); return; }
    try {
      await api.post('/users/me/password', { current_password: currentPassword, new_password: newPassword });
      setCurrentPassword('');
      setNewPassword('');
      setProfileSuccess('Пароль успешно изменён.');
    } catch (error: any) {
      setProfileError(error?.response?.data?.error?.message ?? 'Не удалось изменить пароль. Проверьте текущий пароль.');
    }
  }

  async function login(e: FormEvent) {
    e.preventDefault();
    setError('');

    try {
      const r = await api.post('/auth/login', {
        email,
        password,
      });

      localStorage.setItem(
        'taskflow_access',
        r.data.data.access_token
      );

      localStorage.setItem(
        'taskflow_refresh',
        r.data.data.refresh_token
      );

      setUser(r.data.data.user as UserType);
      nav('/');
    } catch {
      setError('Не удалось войти. Проверьте email и пароль.');
    }
  }

  async function createProject(e: FormEvent) {
    e.preventDefault();
    if (!newProject.trim()) return;
    setCreateError('');
    try {
      const response = await api.post('/projects', {
        name: newProject.trim(),
        description: 'Учебный проект',
      });
      const created = response.data.data as Project;
      setNewProject('');
      setShowCreate(false);
      await qc.invalidateQueries({ queryKey: ['projects'] });
      if (created?.id) setProjectID(created.id);
      setActivePage('board');
    } catch (error: any) {
      setCreateError(error?.response?.data?.error?.message ?? 'Не удалось создать проект. Проверьте права доступа и соединение с сервером.');
    }
  }

  async function createTask(e: FormEvent) {
    e.preventDefault();
    if (!currentProject || !newTask.trim()) return;
    setCreateError('');
    try {
      await api.post(`/projects/${currentProject.id}/tasks`, {
        title: newTask.trim(), status: 'todo', priority: newPriority,
      });
      setNewTask('');
      setNewPriority('medium');
      setShowCreate(false);
      await qc.invalidateQueries({ queryKey: ['tasks'] });
      await qc.invalidateQueries({ queryKey: ['dashboard'] });
    } catch (error: any) {
      setCreateError(error?.response?.data?.error?.message ?? 'Не удалось создать задачу.');
    }
  }

  async function saveProject(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editingProject) return;
    const form = new FormData(e.currentTarget);
    const name = String(form.get('name') ?? '').trim();
    const description = String(form.get('description') ?? '').trim();
    if (!name) {
      setEditError('Название проекта обязательно.');
      return;
    }
    try {
      await api.patch(`/projects/${editingProject.id}`, { name, description });
      setEditingProject(null);
      setEditError('');
      await qc.invalidateQueries({ queryKey: ['projects'] });
    } catch {
      setEditError('Не удалось сохранить проект. Проверьте права доступа.');
    }
  }

  async function deleteProject(project: Project) {
    if (!window.confirm(`Удалить проект «${project.name}» и все его задачи? Это действие нельзя отменить.`)) return;
    try {
      await api.delete(`/projects/${project.id}`);
      if (projectID === project.id) setProjectID(null);
      await qc.invalidateQueries({ queryKey: ['projects'] });
      await qc.invalidateQueries({ queryKey: ['tasks'] });
      await qc.invalidateQueries({ queryKey: ['dashboard'] });
    } catch (error: any) {
      const message = error?.response?.data?.message ?? error?.response?.data?.error?.message ?? error?.message;
      window.alert(`Не удалось удалить проект${message ? `: ${message}` : '. Проверьте права доступа.'}`);
    }
  }

  async function saveTask(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editingTask) return;
    const form = new FormData(e.currentTarget);
    const title = String(form.get('title') ?? '').trim();
    const description = String(form.get('description') ?? '').trim();
    const priority = String(form.get('priority') ?? 'medium') as Task['priority'];
    if (!title) {
      setEditError('Название задачи обязательно.');
      return;
    }
    try {
      await api.patch(`/tasks/${editingTask.id}`, { title, description, priority });
      setEditingTask(null);
      setEditError('');
      await qc.invalidateQueries({ queryKey: ['tasks'] });
      await qc.invalidateQueries({ queryKey: ['dashboard'] });
    } catch {
      setEditError('Не удалось сохранить задачу. Проверьте права доступа.');
    }
  }

  async function deleteTask(task: Task) {
    if (!window.confirm(`Удалить задачу «${task.title}»? Это действие нельзя отменить.`)) return;
    try {
      await api.delete(`/tasks/${task.id}`);
      await qc.invalidateQueries({ queryKey: ['tasks'] });
      await qc.invalidateQueries({ queryKey: ['dashboard'] });
    } catch (error: any) {
      const message = error?.response?.data?.message ?? error?.response?.data?.error?.message ?? error?.message;
      window.alert(`Не удалось удалить задачу${message ? `: ${message}` : '. Проверьте права доступа.'}`);
    }
  }

  async function uploadTaskFile(task: Task, file?: File) {
    if (!file) return;
    const allowed = /\.(png|jpe?g|pdf|txt|docx)$/i.test(file.name);
    if (!allowed) {
      window.alert('Разрешены файлы PNG, JPG, PDF, TXT и DOCX.');
      return;
    }
    setUploadingTaskIds(ids => [...ids, task.id]);
    const form = new FormData();
    form.append('file', file);
    try {
      await api.post(`/tasks/${task.id}/attachment`, form);
      await qc.invalidateQueries({ queryKey: ['tasks'] });
      window.alert('Файл прикреплён к задаче.');
    } catch (error: any) {
      const message = error?.response?.data?.message ?? error?.response?.data?.error?.message;
      window.alert(message ?? 'Не удалось прикрепить файл. Проверьте права доступа и размер файла.');
    } finally {
      setUploadingTaskIds(ids => ids.filter(id => id !== task.id));
    }
  }

  async function downloadTaskFile(task: Task) {
    try {
      const response = await api.get(`/tasks/${task.id}/attachment`, { responseType: 'blob' });
      const disposition = response.headers['content-disposition'] as string | undefined;
      const match = disposition?.match(/filename="?([^";]+)"?/i);
      const filename = match?.[1] ?? `task-${task.id}-attachment`;
      const url = URL.createObjectURL(response.data);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
    } catch {
      window.alert('Не удалось скачать вложение. Возможно, у вас нет доступа к проекту.');
    }
  }

  async function moveTask(
    task: Task,
    status: TaskStatus
  ) {
    if (task.status === status) return;

    await api.patch(`/tasks/${task.id}`, {
      status,
    });

    await qc.invalidateQueries({
      queryKey: ['tasks'],
    });

    await qc.invalidateQueries({
      queryKey: ['dashboard'],
    });
  }

  async function changePriority(
    task: Task,
    priority: Task['priority']
  ) {
    if (task.priority === priority) return;

    await api.patch(`/tasks/${task.id}`, {
      priority,
    });

    await qc.invalidateQueries({
      queryKey: ['tasks'],
    });

    await qc.invalidateQueries({
      queryKey: ['dashboard'],
    });
  }

  function onDragEnd(e: DragEndEvent) {
    if (!e.over) return;

    const task = tasks.find(
      t => String(t.id) === String(e.active.id)
    );

    const status = e.over.id as TaskStatus;

    if (
      task &&
      statuses.some(s => s.key === status)
    ) {
      void moveTask(task, status);
    }
  }

  function handleLogout() {
    setProfileOpen(false);
    logout();
    qc.clear();
    nav('/');
  }

  if (authLoading) {
    return (
      <main className="login-page">
        <div className="login-card">
          <p className="muted">
            Загрузка рабочего пространства…
          </p>
        </div>
      </main>
    );
  }

  if (!user) {
    return (
      <main className="login-page">
        <div className="login-card">
          <div className="brand-mark">TF</div>

          <p className="eyebrow">
            WORKSPACE FOR TEAMS
          </p>

          <h1>С возвращением</h1>

          <p className="muted">
            Войдите в TaskFlow, чтобы продолжить работу.
          </p>

          <form onSubmit={login} className="form-stack">
            <label>
              Email
              <input
                type="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
              />
            </label>

            <label>
              Пароль
              <input
                type="password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                required
              />
            </label>

            {error && (
              <div className="error-box">
                {error}
              </div>
            )}

            <button className="primary full">
              Войти в рабочее пространство
            </button>
          </form>

          <p className="hint">
            Демо: lead@taskflow.local · ChangeMe123!
          </p>

          <p className="hint">
            Также доступны admin@taskflow.local,
            student@taskflow.local и teacher@taskflow.local.
          </p>
        </div>
      </main>
    );
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <Link to="/" className="brand">
          <span className="brand-mark small">TF</span>
          <span>taskflow</span>
        </Link>

        <div className="workspace">
          <div className="avatar">
            {user.name.slice(0, 1).toUpperCase()}
          </div>

          <div>
            <strong>{user.name}</strong>
            <small>{user.role}</small>
          </div>

          <span className="online-dot" />
        </div>

        <p className="nav-caption">WORKSPACE</p>

        <button
          className={`nav-item ${
            activePage === 'board' ? 'selected' : ''
          }`}
          onClick={() => setActivePage('board')}
        >
          <FolderKanban size={18} />
          Канбан-доска
        </button>

        <button
          className={`nav-item ${
            activePage === 'dashboard' ? 'selected' : ''
          }`}
          onClick={() => setActivePage('dashboard')}
        >
          <BarChart3 size={18} />
          Аналитика
        </button>

        <button
          className={`nav-item ${
            activePage === 'activity' ? 'selected' : ''
          }`}
          onClick={() => setActivePage('activity')}
        >
          <Activity size={18} />
          Активность
        </button>

        <p className="nav-caption">
          PROJECTS

          <button
            className="icon-button tiny"
            onClick={() => { setCreateKind('project'); setCreateError(''); setShowCreate(true); }}
            aria-label="Создать проект"
          >
            <Plus size={14} />
          </button>
        </p>

        <div className="project-list">
          {projects.map(p => (
            <div className={`project-row ${currentProject?.id === p.id ? 'active' : ''}`} key={p.id}>
              <button
                className={`project-link ${currentProject?.id === p.id ? 'active' : ''}`}
                onClick={() => setProjectID(p.id)}
              >
                <span className="project-color" />
                <span className="project-name">{p.name}</span>
              </button>
              <button className="row-action" aria-label={`Редактировать проект ${p.name}`} title="Редактировать проект" onClick={() => { setEditError(''); setEditingProject(p); }}><Pencil size={13} /></button>
              <button className="row-action danger" aria-label={`Удалить проект ${p.name}`} title="Удалить проект" onClick={() => void deleteProject(p)}><Trash2 size={13} /></button>
            </div>
          ))}
        </div>

        <button
          className="nav-item"
          onClick={() => { setCreateKind('project'); setCreateError(''); setShowCreate(true); }}
        >
          <Plus size={18} />
          Новый проект
        </button>

        <div className="sidebar-bottom">
          <button
            className="nav-item"
            onClick={() => setDark(!dark)}
          >
            {dark ? <Sun size={18} /> : <Moon size={18} />}
            {dark ? 'Светлая тема' : 'Тёмная тема'}
          </button>

          {user.role === 'admin' && (
            <button
              className="nav-item"
              onClick={() => setActivePage('admin')}
            >
              <ShieldCheck size={18} />
              Администрирование
            </button>
          )}

          <button
            className="nav-item"
            onClick={handleLogout}
          >
            <LogOut size={18} />
            Выйти
          </button>
        </div>
      </aside>

      <main className="main-area">
        <header className="topbar">
          <div className="breadcrumbs">
            Workspace
            <span>/</span>
            {currentProject?.name ?? 'Проекты'}
          </div>

          <div className="top-actions">
            <button
              className="icon-button"
              title="Помощь"
              aria-label="Помощь"
              onClick={() => {
                setActiveModal('help');
                setProfileOpen(false);
              }}
            >
              <CircleHelp size={19} />
            </button>

            <button
              className="icon-button"
              title="Настройки"
              aria-label="Настройки"
              onClick={() => {
                setActiveModal('settings');
                setProfileOpen(false);
              }}
            >
              <Settings size={19} />
            </button>

            <div style={{ position: 'relative' }}>
              <button
                className="avatar"
                type="button"
                aria-label="Профиль пользователя"
                aria-expanded={profileOpen}
                onClick={() => setProfileOpen(!profileOpen)}
                style={{
                  border: 0,
                  cursor: 'pointer',
                  width: 44,
                  height: 44,
                  fontSize: 19,
                }}
              >
                {user.name.slice(0, 1).toUpperCase()}
              </button>

              {profileOpen && (
                <div
                  style={{
                    position: 'absolute',
                    right: 0,
                    top: '54px',
                    zIndex: 30,
                    width: 220,
                    padding: 12,
                    background: 'var(--surface)',
                    border: '1px solid var(--border)',
                    borderRadius: 12,
                    boxShadow: 'var(--shadow)',
                  }}
                >
                  <div
                    style={{
                      padding: '8px 6px 12px',
                      borderBottom: '1px solid var(--border)',
                      marginBottom: 8,
                    }}
                  >
                    <strong
                      style={{
                        display: 'block',
                        color: 'var(--text)',
                        fontSize: 13,
                      }}
                    >
                      {user.name}
                    </strong>

                    <span
                      style={{
                        color: 'var(--muted)',
                        fontSize: 11,
                      }}
                    >
                      {user.email}
                    </span>

                    <p
                      className="muted"
                      style={{ marginTop: 5 }}
                    >
                      Роль: {user.role}
                    </p>
                  </div>

                  <button
                    className="nav-item"
                    style={{ margin: 0 }}
                    onClick={() => {
                      setProfileOpen(false);
                      setActiveModal('settings');
                    }}
                  >
                    <User size={16} />
                    Профиль и настройки
                  </button>

                  <button
                    className="nav-item"
                    style={{ margin: 0 }}
                    onClick={handleLogout}
                  >
                    <LogOut size={16} />
                    Выйти из аккаунта
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>

        <div className="content">
          <div className="page-heading">
            <div>
              <p className="eyebrow">
                PROJECT OVERVIEW
              </p>

              <h1>
                {activePage === 'dashboard'
                  ? 'Аналитика'
                  : activePage === 'admin'
                    ? 'Администрирование'
                    : activePage === 'activity'
                      ? 'История активности'
                      : currentProject?.name ?? 'Ваши проекты'}
              </h1>

              <p className="muted">
                {activePage === 'board'
                  ? 'Планируйте, отслеживайте и выполняйте работу вместе.'
                  : 'Следите за прогрессом команды и состоянием проекта.'}
              </p>
            </div>

            <button
              className="primary"
              onClick={() => { setCreateKind(activePage === 'board' ? 'task' : 'project'); setCreateError(''); setShowCreate(true); }}
            >
              <Plus size={17} />
              {activePage === 'board' ? 'Новая задача' : 'Создать'}
            </button>
          </div>

          {activePage === 'admin' && user.role === 'admin' ? (
            <AdminUsers />
          ) : activePage === 'dashboard' ? (
            <Dashboard data={dashboard} tasks={tasks} />
          ) : activePage === 'activity' ? (
            <ActivityPage project={currentProject} />
          ) : (
            <>
              <div className="stats-grid">
                <Stat
                  label="Всего задач"
                  value={dashboard?.total_tasks ?? tasks.length}
                  icon={<ClipboardList />}
                  tone="violet"
                />

                <Stat
                  label="Завершено"
                  value={
                    dashboard?.completed_tasks ??
                    tasks.filter(t => t.status === 'done').length
                  }
                  icon={<CheckCircle2 />}
                  tone="green"
                />

                <Stat
                  label="Просрочено"
                  value={dashboard?.overdue_tasks ?? 0}
                  icon={<Activity />}
                  tone="orange"
                />

                <Stat
                  label="Выполнено"
                  value={`${Math.round(
                    dashboard?.completion_percent ?? 0
                  )}%`}
                  icon={<BarChart3 />}
                  tone="blue"
                />
              </div>

              <div className="toolbar task-toolbar">
                <div className="search-box">
                  <Search size={17} />
                  <input placeholder="Поиск задач..." value={search} onChange={e => { setSearch(e.target.value); setTaskPage(1); }} />
                </div>
                <select value={filter} onChange={e => { setFilter(e.target.value); setTaskPage(1); }} aria-label="Фильтр по приоритету">
                  <option value="">Все приоритеты</option>
                  {priorities.map(p => <option key={p.key} value={p.key}>{p.label}</option>)}
                </select>
                <select value={statusFilter} onChange={e => { setStatusFilter(e.target.value); setTaskPage(1); }} aria-label="Фильтр по статусу">
                  <option value="">Все статусы</option>
                  {statuses.map(s => <option key={s.key} value={s.key}>{s.label}</option>)}
                </select>
                <select value={assigneeFilter} onChange={e => { setAssigneeFilter(e.target.value); setTaskPage(1); }} aria-label="Фильтр по исполнителю">
                  <option value="">Все исполнители</option>
                  <option value="unassigned">Без исполнителя</option>
                  {(membersQuery.data ?? []).map(member => <option key={member.id} value={member.id}>{member.name || member.email}</option>)}
                </select>
                <select value={sort} onChange={e => { setSort(e.target.value); setTaskPage(1); }} aria-label="Сортировка задач">
                  <option value="created_at">По дате создания</option>
                  <option value="updated_at">По обновлению</option>
                  <option value="due_date">По сроку</option>
                  <option value="priority">По приоритету</option>
                  <option value="title">По названию</option>
                </select>
                <select value={sortOrder} onChange={e => { setSortOrder(e.target.value as 'asc' | 'desc'); setTaskPage(1); }} aria-label="Направление сортировки">
                  <option value="desc">По убыванию</option>
                  <option value="asc">По возрастанию</option>
                </select>
                <span className="task-count">{pagination?.total ?? tasks.length} задач</span>
              </div>

              {!currentProject ? (
                <div className="empty-state">
                  <FolderKanban size={32} />

                  <h3>Создайте первый проект</h3>

                  <p>Задачи будут отображаться здесь.</p>

                  <button
                    className="primary"
                    onClick={() => { setCreateKind('project'); setCreateError(''); setShowCreate(true); }}
                  >
                    Создать проект
                  </button>
                </div>
              ) : tasksQuery.isLoading ? (
                <p className="muted">
                  Загружаем задачи…
                </p>
              ) : tasksQuery.isError ? (
                <div className="error-box">
                  Не удалось загрузить задачи. Проверьте подключение к API.
                </div>
              ) : (
                <DndContext onDragEnd={onDragEnd}>
                  <div className="kanban">
                    {statuses.map(status => (
                      <KanbanColumn
                        key={status.key}
                        status={status.key}
                        label={status.label}
                        tasks={tasks.filter(
                          t => t.status === status.key
                        )}
                        onMove={moveTask}
                        onPriorityChange={changePriority}
                        onEdit={task => { setEditError(''); setEditingTask(task); }}
                        onDelete={deleteTask}
                        onUpload={uploadTaskFile}
                        onDownload={downloadTaskFile}
                        uploadingTaskIds={uploadingTaskIds}
                      />
                    ))}
                  </div>
                </DndContext>
              )}
              {currentProject && !tasksQuery.isLoading && !tasksQuery.isError && (
                <div className="pagination-bar">
                  <label>На странице
                    <select value={pageSize} onChange={e => { setPageSize(Number(e.target.value)); setTaskPage(1); }} aria-label="Количество задач на странице">
                      <option value={10}>10</option><option value={20}>20</option><option value={50}>50</option><option value={100}>100</option>
                    </select>
                  </label>
                  <span>Страница {pagination?.page ?? taskPage} из {totalPages}</span>
                  <div className="pagination-actions">
                    <button type="button" className="secondary" disabled={taskPage <= 1} onClick={() => setTaskPage(p => Math.max(1, p - 1))}>Назад</button>
                    <button type="button" className="secondary" disabled={taskPage >= totalPages} onClick={() => setTaskPage(p => Math.min(totalPages, p + 1))}>Далее</button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </main>

      {showCreate && (
        <div
          className="modal-backdrop"
          onMouseDown={e => {
            if (e.target === e.currentTarget) {
              setShowCreate(false);
              setCreateError('');
            }
          }}
        >
          <div className="modal">
            <button
              className="modal-close"
              type="button"
              onClick={() => setShowCreate(false)}
              aria-label="Закрыть"
            >
              ×
            </button>

            <p className="eyebrow">
              QUICK CREATE
            </p>

            <h2>
              {createKind === 'task' ? 'Создать задачу' : 'Создать проект'}
            </h2>

            <form
              className="form-stack"
              onSubmit={
                createKind === 'task' ? createTask : createProject
              }
            >
              <label>
                {createKind === 'task' ? 'Название задачи' : 'Название проекта'}

                <input
                  autoFocus
                  value={createKind === 'task' ? newTask : newProject}
                  onChange={e =>
                    createKind === 'task'
                      ? setNewTask(e.target.value)
                      : setNewProject(e.target.value)
                  }
                  placeholder={
                    createKind === 'task'
                      ? 'Например, подготовить API'
                      : 'Например, дипломный проект'
                  }
                  required
                />
              </label>

              {createKind === 'task' && currentProject && (
                <label>
                  Приоритет задачи

                  <select
                    value={newPriority}
                    onChange={e =>
                      setNewPriority(
                        e.target.value as Task['priority']
                      )
                    }
                  >
                    {priorities.map(p => (
                      <option key={p.key} value={p.key}>
                        {p.label}
                      </option>
                    ))}
                  </select>
                </label>
              )}

              {createError && <div className="error-box" role="alert">{createError}</div>}
              <button className="primary full">
                Создать
              </button>
            </form>
          </div>
        </div>
      )}

      {editingProject && (
        <div className="modal-backdrop" onMouseDown={e => { if (e.target === e.currentTarget) setEditingProject(null); }}>
          <div className="modal">
            <button className="modal-close" type="button" onClick={() => setEditingProject(null)} aria-label="Закрыть"><X size={19} /></button>
            <p className="eyebrow">PROJECT SETTINGS</p>
            <h2>Редактировать проект</h2>
            <form className="form-stack" onSubmit={saveProject}>
              <label>Название проекта<input name="name" defaultValue={editingProject.name} maxLength={120} required /></label>
              <label>Описание<textarea name="description" defaultValue={editingProject.description} rows={4} maxLength={2000} /></label>
              {editError && <div className="error-box" role="alert">{editError}</div>}
              <button className="primary full" type="submit">Сохранить изменения</button>
            </form>
          </div>
        </div>
      )}

      {editingTask && (
        <div className="modal-backdrop" onMouseDown={e => { if (e.target === e.currentTarget) setEditingTask(null); }}>
          <div className="modal">
            <button className="modal-close" type="button" onClick={() => setEditingTask(null)} aria-label="Закрыть"><X size={19} /></button>
            <p className="eyebrow">TASK SETTINGS · TF-{editingTask.id}</p>
            <h2>Редактировать задачу</h2>
            <form className="form-stack" onSubmit={saveTask}>
              <label>Название задачи<input name="title" defaultValue={editingTask.title} maxLength={200} required /></label>
              <label>Описание<textarea name="description" defaultValue={editingTask.description} rows={4} maxLength={5000} /></label>
              <label>Приоритет<select name="priority" defaultValue={editingTask.priority}>{priorities.map(p => <option key={p.key} value={p.key}>{p.label}</option>)}</select></label>
              {editError && <div className="error-box" role="alert">{editError}</div>}
              <button className="primary full" type="submit">Сохранить изменения</button>
            </form>
          </div>
        </div>
      )}

      {activeModal && (
        <div
          className="modal-backdrop"
          onMouseDown={e => {
            if (e.target === e.currentTarget) {
              setActiveModal(null);
            }
          }}
        >
          <div className="modal">
            <button
              className="modal-close"
              type="button"
              onClick={() => setActiveModal(null)}
              aria-label="Закрыть"
            >
              <X size={19} />
            </button>

            {activeModal === 'help' ? (
              <>
                <p className="eyebrow">TASKFLOW HELP</p>

                <h2>Помощь</h2>

                <div className="form-stack">
                  <div>
                    <strong>Канбан-доска</strong>
                    <p className="muted">
                      Создавайте задачи и перемещайте их между
                      колонками перетаскиванием или изменением статуса.
                    </p>
                  </div>

                  <div>
                    <strong>Приоритеты</strong>
                    <p className="muted">
                      У каждой задачи есть четыре уровня:
                      низкий, средний, высокий и критический.
                      Приоритет можно изменить в карточке задачи.
                    </p>
                  </div>

                  <div>
                    <strong>Аналитика</strong>
                    <p className="muted">
                      Следите за количеством задач и распределением
                      работы по статусам.
                    </p>
                  </div>

                  <div>
                    <strong>Проекты</strong>
                    <p className="muted">
                      Создавайте отдельные проекты для разных
                      учебных команд и выбирайте их в боковой панели.
                    </p>
                  </div>
                </div>
              </>
            ) : (
              <>
                <p className="eyebrow">USER SETTINGS</p>

                <h2>Настройки</h2>

                <div className="form-stack">
                  <form className="form-stack profile-form" onSubmit={saveProfile}>
                    <strong>Профиль</strong>
                    <label>Имя
                      <input value={profileName} onChange={e => setProfileName(e.target.value)} maxLength={120} required />
                    </label>
                    <p className="muted">Email: {user.email}</p>
                    <p className="muted">Роль: {user.role}</p>
                    <button className="primary" type="submit">Сохранить имя</button>
                  </form>

                  <form className="form-stack profile-form" onSubmit={changeOwnPassword}>
                    <strong>Изменение пароля</strong>
                    <label>Текущий пароль
                      <input type="password" autoComplete="current-password" value={currentPassword} onChange={e => setCurrentPassword(e.target.value)} required />
                    </label>
                    <label>Новый пароль
                      <input type="password" autoComplete="new-password" minLength={8} value={newPassword} onChange={e => setNewPassword(e.target.value)} required />
                    </label>
                    <button className="secondary" type="submit">Изменить пароль</button>
                  </form>
                  {profileError && <div className="error-box">{profileError}</div>}
                  {profileSuccess && <div className="success-box">{profileSuccess}</div>}

                  <div>
                    <strong>Внешний вид</strong>

                    <p className="muted">
                      Настройте оформление рабочего пространства.
                    </p>

                    <button
                      type="button"
                      className="secondary"
                      onClick={() => setDark(!dark)}
                    >
                      {dark ? (
                        <>
                          <Sun size={15} />
                          Светлая тема
                        </>
                      ) : (
                        <>
                          <Moon size={15} />
                          Тёмная тема
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function Stat({
  label,
  value,
  icon,
  tone,
}: {
  label: string;
  value: string | number;
  icon: ReactNode;
  tone: string;
}) {
  return (
    <div className="stat-card">
      <div className={`stat-icon ${tone}`}>
        {icon}
      </div>

      <div>
        <p>{label}</p>
        <strong>{value}</strong>
      </div>

      <span className="stat-decoration" />
    </div>
  );
}

function KanbanColumn({
  status,
  label,
  tasks,
  onMove,
  onPriorityChange,
  onEdit,
  onDelete,
  onUpload,
  onDownload,
  uploadingTaskIds,
}: {
  status: TaskStatus;
  label: string;
  tasks: Task[];
  onMove: (task: Task, status: TaskStatus) => void;
  onPriorityChange: (task: Task, priority: Task['priority']) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onUpload: (task: Task, file?: File) => void;
  onDownload: (task: Task) => void;
  uploadingTaskIds: number[];
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: status,
  });

  return (
    <section
      ref={setNodeRef}
      className={`kanban-column ${
        isOver ? 'drop-target' : ''
      }`}
    >
      <div className="column-heading">
        <span className={`status-dot ${status}`} />
        <h3>{label}</h3>
        <span className="column-count">
          {tasks.length}
        </span>
      </div>

      <div className="task-stack">
        {tasks.map(task => (
          <TaskCard
            key={task.id}
            task={task}
            onMove={onMove}
            onPriorityChange={onPriorityChange}
            onEdit={onEdit}
            onDelete={onDelete}
            onUpload={onUpload}
            onDownload={onDownload}
            isUploading={uploadingTaskIds.includes(task.id)}
          />
        ))}
      </div>
    </section>
  );
}

function TaskCard({
  task,
  onMove,
  onPriorityChange,
  onEdit,
  onDelete,
  onUpload,
  onDownload,
  isUploading,
}: {
  task: Task;
  onMove: (task: Task, status: TaskStatus) => void;
  onPriorityChange: (task: Task, priority: Task['priority']) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onUpload: (task: Task, file?: File) => void;
  onDownload: (task: Task) => void;
  isUploading: boolean;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    isDragging,
  } = useDraggable({
    id: String(task.id),
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
        opacity: isDragging ? 0.6 : 1,
      }
    : undefined;

  return (
    <article
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      className="task-card"
    >
      <div className="task-card-top">
        <span className={`priority ${task.priority}`}>
          {priorityLabel(task.priority)}
        </span>
        <span className="task-id">TF-{task.id}</span>
        <div className="task-actions">
          <button type="button" className="row-action" aria-label={`Редактировать задачу ${task.title}`} title="Редактировать задачу" onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onEdit(task); }}><Pencil size={13} /></button>
          <button type="button" className="row-action danger" aria-label={`Удалить задачу ${task.title}`} title="Удалить задачу" onPointerDown={e => e.stopPropagation()} onClick={e => { e.stopPropagation(); onDelete(task); }}><Trash2 size={13} /></button>
        </div>
      </div>

      <h4>{task.title}</h4>

      {task.description && (
        <p>{task.description}</p>
      )}

      <div className="task-attachments" onPointerDown={e => e.stopPropagation()} onClick={e => e.stopPropagation()}>
        {task.attachment_path && (
          <button type="button" className="attachment-download" onClick={() => onDownload(task)}>
            <Download size={12} /> Скачать вложение
          </button>
        )}
        <label className="attachment-upload">
          <UploadIcon size={12} /> {isUploading ? 'Загрузка…' : task.attachment_path ? 'Заменить файл' : 'Прикрепить файл'}
          <input
            type="file"
            accept=".png,.jpg,.jpeg,.pdf,.txt,.docx"
            disabled={isUploading}
            onChange={e => {
              const file = e.currentTarget.files?.[0];
              if (file) onUpload(task, file);
              e.currentTarget.value = '';
            }}
          />
        </label>
      </div>

      <div className="task-card-bottom">
        <span className="task-date">
          {task.due_date
            ? new Date(task.due_date).toLocaleDateString('ru-RU')
            : 'Без срока'}
        </span>

        <select
          aria-label={`Статус задачи ${task.title}`}
          value={task.status}
          onChange={e =>
            onMove(task, e.target.value as TaskStatus)
          }
          onPointerDown={e => e.stopPropagation()}
          onClick={e => e.stopPropagation()}
        >
          {statuses.map(s => (
            <option key={s.key} value={s.key}>
              {s.label}
            </option>
          ))}
        </select>
      </div>

      <div
        style={{
          marginTop: 10,
          display: 'flex',
          alignItems: 'center',
          gap: 7,
        }}
      >
        <span
          style={{
            fontSize: 10,
            color: 'var(--muted)',
            whiteSpace: 'nowrap',
          }}
        >
          Приоритет:
        </span>

        <select
          aria-label={`Приоритет задачи ${task.title}`}
          value={task.priority}
          onChange={e =>
            onPriorityChange(
              task,
              e.target.value as Task['priority']
            )
          }
          onPointerDown={e => e.stopPropagation()}
          onClick={e => e.stopPropagation()}
          style={{
            flex: 1,
            minWidth: 0,
            padding: '5px 7px',
            background: 'var(--surface2)',
            border: '1px solid var(--border)',
            borderRadius: 7,
            color: 'var(--text)',
            fontSize: 10,
          }}
        >
          {priorities.map(p => (
            <option key={p.key} value={p.key}>
              {p.label}
            </option>
          ))}
        </select>
      </div>
    </article>
  );
}

type DashboardData = {
  total_tasks?: number;
  completed_tasks?: number;
  overdue_tasks?: number;
  completion_percent?: number;
};

function Dashboard({
  data,
  tasks,
}: {
  data: DashboardData | undefined;
  tasks: Task[];
}) {
  const chart = useMemo(
    () =>
      statuses.map(s => ({
        name: s.label,
        count: tasks.filter(t => t.status === s.key).length,
      })),
    [tasks]
  );

  return (
    <>
      <div className="stats-grid">
        <Stat
          label="Всего задач"
          value={data?.total_tasks ?? tasks.length}
          icon={<ClipboardList />}
          tone="violet"
        />

        <Stat
          label="Завершено"
          value={data?.completed_tasks ?? 0}
          icon={<CheckCircle2 />}
          tone="green"
        />

        <Stat
          label="Просрочено"
          value={data?.overdue_tasks ?? 0}
          icon={<Activity />}
          tone="orange"
        />

        <Stat
          label="Выполнено"
          value={`${Math.round(
            data?.completion_percent ?? 0
          )}%`}
          icon={<BarChart3 />}
          tone="blue"
        />
      </div>

      <section className="panel chart-panel">
        <div className="panel-heading">
          <div>
            <h2>Задачи по статусам</h2>
            <p className="muted">
              Текущее распределение задач проекта
            </p>
          </div>

          <span className="pill">
            Обновляется автоматически
          </span>
        </div>

        <div className="chart">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chart}>
              <CartesianGrid
                strokeDasharray="3 3"
                vertical={false}
              />

              <XAxis dataKey="name" />
              <YAxis allowDecimals={false} />
              <Tooltip />

              <Bar
                dataKey="count"
                fill="#8b5cf6"
                radius={[6, 6, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </section>
    </>
  );
}

function AdminUsers() {
  const q = useQuery({
    queryKey: ['admin-users'],
    queryFn: async () =>
      (await api.get('/admin/users')).data.data as UserType[],
  });

  return (
    <section className="panel">
      <div className="panel-heading">
        <div>
          <h2>Пользователи</h2>
          <p className="muted">
            Управление ролями и доступом
          </p>
        </div>

        <Users />
      </div>

      {q.isLoading ? (
        'Загрузка…'
      ) : q.isError ? (
        <p className="muted">
          Не удалось загрузить пользователей.
        </p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Пользователь</th>
                <th>Email</th>
                <th>Роль</th>
                <th>Статус</th>
                <th>Действия</th>
              </tr>
            </thead>

            <tbody>
              {q.data?.map(u => {
                const blocked = (
                  u as UserType & { is_blocked?: boolean }
                ).is_blocked;

                return (
                  <tr key={u.id}>
                    <td>{u.name}</td>
                    <td>{u.email}</td>

                    <td>
                      <select
                        value={u.role}
                        onChange={async e => {
                          await api.patch(
                            `/admin/users/${u.id}/role`,
                            { role: e.target.value }
                          );

                          await q.refetch();
                        }}
                      >
                        {[
                          'admin',
                          'lead',
                          'participant',
                          'teacher',
                        ].map(r => (
                          <option key={r} value={r}>
                            {r}
                          </option>
                        ))}
                      </select>
                    </td>

                    <td>
                      {blocked
                        ? 'Заблокирован'
                        : 'Активен'}
                    </td>

                    <td>
                      <button
                        className="secondary small-button"
                        onClick={async () => {
                          await api.patch(
                            `/admin/users/${u.id}/status`,
                            { blocked: !blocked }
                          );

                          await q.refetch();
                        }}
                      >
                        Переключить блокировку
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function ActivityPage({ project }: { project: Project | undefined }) {
  type HistoryItem = {
    id: number; task_id: number; task_title: string; actor_id: number;
    actor_name: string; field: string; old_value: string; new_value: string; created_at: string;
  };
  const q = useQuery({
    queryKey: ['activity', project?.id],
    queryFn: async () => (await api.get(`/projects/${project!.id}/history`)).data.data as HistoryItem[],
    enabled: !!project,
  });
  const fieldLabels: Record<string, string> = {
    created: 'создал задачу', status: 'изменил статус', title: 'изменил название',
    description: 'изменил описание', priority: 'изменил приоритет',
    assignee_id: 'изменил исполнителя', due_date: 'изменил срок',
  };
  const valueLabel = (field: string, value: string) => {
    if (field === 'status') return statuses.find(item => item.key === value)?.label ?? value;
    if (field === 'priority') return priorities.find(item => item.key === value)?.label ?? value;
    return value || '—';
  };
  return (
    <section className="panel">
      <div className="panel-heading">
        <div><h2>История изменений</h2><p className="muted">Последние действия участников в проекте</p></div>
        <Activity />
      </div>
      {!project ? <p className="muted">Сначала выберите проект.</p> : q.isLoading ? <p className="muted">Загрузка истории…</p> : q.isError ? <div className="error-box">Не удалось загрузить историю проекта.</div> : q.data?.length ? q.data.map(item => (
        <div className="activity-row" key={item.id}>
          <span className="status-dot in_progress" />
          <div>
            <strong>{item.actor_name} {fieldLabels[item.field] ?? 'изменил задачу'} «{item.task_title}»</strong>
            <p className="muted">TF-{item.task_id} · {new Date(item.created_at).toLocaleString()}</p>
            {item.old_value !== item.new_value && <p className="muted">{valueLabel(item.field, item.old_value)} → {valueLabel(item.field, item.new_value)}</p>}
          </div>
        </div>
      )) : <p className="muted">История пока пуста. Изменения задач будут отображаться здесь.</p>}
    </section>
  );
}

export default App;