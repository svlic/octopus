# Octopus Admin UI

This directory contains the Next.js admin UI embedded in the Octopus Go binary for production releases.

Use pnpm for frontend development and builds:

```bash
pnpm install
NEXT_PUBLIC_API_BASE_URL="http://127.0.0.1:8080" pnpm run dev
```

The development server runs at <http://localhost:3000>. Start the Go backend separately on port 8080.

For production builds, deployment, and complete project setup, see the [repository README](../README.md).
