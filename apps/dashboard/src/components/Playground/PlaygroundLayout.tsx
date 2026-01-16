function PlaygroundLayout({ children }: { children: React.ReactNode }) {
  return <div className="grid grid-cols-1 lg:grid-cols-[minmax(320px,400px)_1fr] h-full">{children}</div>
}

function PlaygroundLayoutSidebar({ children }: { children: React.ReactNode }) {
  return <div className="overflow-auto bg-sidebar/20 p-4 border-r border-border hidden lg:block">{children}</div>
}

function PlaygroundLayoutContent({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-auto bg-[radial-gradient(hsl(var(--border))_1px,transparent_1px)] [background-size:12px_12px] flex items-center justify-center p-5">
      {children}
    </div>
  )
}

export { PlaygroundLayout, PlaygroundLayoutContent, PlaygroundLayoutSidebar }
