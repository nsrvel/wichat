# WiChat web-client — code templates

Use with `.cursor/rules/wichat-web-layout.mdc` and `wichat-web-data.mdc`. Paths assume `src/`.

## Thin route

```tsx
// app/(auth)/login/page.tsx
import { LoginPage } from '@/features/auth/login';

export default function Page() {
  return <LoginPage />;
}
```

## `get-query-client.ts` (sketch)

```ts
import { QueryClient, isServer } from '@tanstack/react-query';

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { staleTime: 60_000 } },
  });
}

let browserClient: QueryClient | undefined;

export function getQueryClient() {
  if (isServer) return makeQueryClient();
  browserClient ??= makeQueryClient();
  return browserClient;
}
```

## `api-client` (sketch)

```ts
import { env } from '@/env';
import { ApiError } from '@/lib/api-error';

const base = `${env.NEXT_PUBLIC_API_URL}/api/v1`;

export async function post<T>(
  path: string,
  schema: { parse: (u: unknown) => T },
  body: unknown,
): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const json = await res.json();
  if (!res.ok) throw ApiError.fromEnvelope(json);
  return schema.parse(json.data);
}
```

## Service barrel

```ts
// services/auth/index.ts
export * as authKey from './auth.key';
export * as authQuery from './auth.query';
export * as authMutation from './auth.mutation';
```

## Form + mutation

```tsx
'use client';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { LoginRequestSchema } from '@/schemas/auth';
import { useLoginMutation } from '@/services/auth/auth.mutation';

export function LoginForm() {
  const form = useForm({
    resolver: zodResolver(LoginRequestSchema),
    defaultValues: { email: '', password: '' },
  });
  const login = useLoginMutation();

  return (
    <form onSubmit={form.handleSubmit((v) => login.mutate(v))}>
      {/* shadcn FormField … */}
    </form>
  );
}
```

## Anti-patterns

```tsx
// ❌ fetch in components/ui/button.tsx
// ❌ createServerFn + postgres in .service.ts
// ❌ components/auth/login-form.tsx when only auth uses it — use features/auth/login/
```

Reference (TanStack Start, not copy runtime): `references/react-tanstack/src/services/auth/`.
