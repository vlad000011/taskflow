
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';
import { useAuth, type User } from './stores/auth';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
}));

// Изолируем тесты от реального API и браузерного хранилища.
vi.mock('./lib/api', () => ({ api: { get: mocks.get } }));

function createStorageMock() {
  const values = new Map<string, string>();

  return {
    getItem: vi.fn((key: string) => values.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) =>
      values.set(key, String(value))
    ),
    removeItem: vi.fn((key: string) => values.delete(key)),
    clear: vi.fn(() => values.clear()),
  };
}

function renderApp(user: User | null) {
  useAuth.setState({ user });

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </MemoryRouter>
  );
}

const account = (role: string): User => ({
  id: 1,
  email: `${role}@taskflow.local`,
  name: 'Тестовый пользователь',
  role,
});

beforeEach(() => {
  vi.stubGlobal('localStorage', createStorageMock());

  mocks.get.mockReset();
  mocks.get.mockResolvedValue({ data: { data: [] } });

  useAuth.setState({ user: null });
});

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
  vi.unstubAllGlobals();
});

describe('TaskFlow React component rendering', () => {
  it('renders login form for unauthenticated visitors', async () => {
    renderApp(null);

    expect(
      await screen.findByRole('heading', { name: 'С возвращением' })
    ).toBeInTheDocument();

    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Пароль')).toBeInTheDocument();

    expect(
      screen.getByRole('button', {
        name: 'Войти в рабочее пространство',
      })
    ).toBeInTheDocument();
  });

  it('shows administration navigation for an administrator', async () => {
    renderApp(account('admin'));

    expect(
      await screen.findByRole('button', {
        name: 'Администрирование',
      })
    ).toBeInTheDocument();

    expect(
      screen.getByRole('button', { name: 'Канбан-доска' })
    ).toBeInTheDocument();
  });

  it('hides administration navigation from a teacher', async () => {
    renderApp(account('teacher'));

    expect(
      await screen.findByRole('button', { name: 'Канбан-доска' })
    ).toBeInTheDocument();

    expect(
      screen.queryByRole('button', { name: 'Администрирование' })
    ).not.toBeInTheDocument();
  });
});
