// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>{children}</body>
    </html>
  )
}
