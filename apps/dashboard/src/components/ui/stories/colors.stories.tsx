import type { Meta, StoryObj } from '@storybook/react'
import { useState, useCallback } from 'react'
import { MoreVertical, Check } from 'lucide-react'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '../dropdown-menu'
import { Separator } from '../separator'

function hslToHex(hsl: string): string {
  const parts = hsl.trim().split(/\s+/)
  if (parts.length < 3) return '#000000'
  const h = parseFloat(parts[0])
  const s = parseFloat(parts[1]) / 100
  const l = parseFloat(parts[2]) / 100

  const a = s * Math.min(l, 1 - l)
  const f = (n: number) => {
    const k = (n + h / 30) % 12
    const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1)
    return Math.round(255 * color)
      .toString(16)
      .padStart(2, '0')
  }
  return `#${f(0)}${f(8)}${f(4)}`
}

function ColorSwatch({ variable, label }: { variable: string; label: string }) {
  const [copied, setCopied] = useState<string | null>(null)

  const copyToClipboard = useCallback((text: string, type: string) => {
    navigator.clipboard.writeText(text)
    setCopied(type)
    setTimeout(() => setCopied(null), 1500)
  }, [])

  const hslValue =
    typeof window !== 'undefined' ? getComputedStyle(document.documentElement).getPropertyValue(variable).trim() : ''
  const hex = hslToHex(hslValue)

  return (
    <div>
      <div
        className="relative w-[100px] aspect-square rounded-md border border-border"
        style={{ backgroundColor: `hsl(var(${variable}))` }}
      >
        <div style={{ position: 'absolute', top: 4, right: 4, zIndex: 10 }}>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                style={{
                  padding: 2,
                  borderRadius: 4,
                  backgroundColor: 'rgba(255,255,255,0.8)',
                  border: '1px solid #e5e5e5',
                  boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <MoreVertical className="size-3.5" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-[140px]">
              <DropdownMenuItem onClick={() => copyToClipboard(hex, 'hex')}>
                {copied === 'hex' ? <Check className="size-3.5" /> : null}
                Copy hex
                <span className="ml-auto text-xs text-muted-foreground">{hex}</span>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => copyToClipboard(`var(${variable})`, 'var')}>
                {copied === 'var' ? <Check className="size-3.5" /> : null}
                Copy variable
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
      <div className="mt-1.5 text-xs leading-tight">
        <div className="font-medium truncate w-[100px]">{label}</div>
        <div className="text-muted-foreground truncate w-[100px]">{variable}</div>
      </div>
    </div>
  )
}

const colorGroups = [
  {
    name: 'Base',
    colors: [
      { variable: '--background', label: 'Background' },
      { variable: '--foreground', label: 'Foreground' },
      { variable: '--card', label: 'Card' },
      { variable: '--card-foreground', label: 'Card FG' },
      { variable: '--popover', label: 'Popover' },
      { variable: '--popover-foreground', label: 'Popover FG' },
    ],
  },
  {
    name: 'Brand',
    colors: [
      { variable: '--primary', label: 'Primary' },
      { variable: '--primary-foreground', label: 'Primary FG' },
      { variable: '--secondary', label: 'Secondary' },
      { variable: '--secondary-foreground', label: 'Secondary FG' },
      { variable: '--muted', label: 'Muted' },
      { variable: '--muted-foreground', label: 'Muted FG' },
      { variable: '--accent', label: 'Accent' },
      { variable: '--accent-foreground', label: 'Accent FG' },
    ],
  },
  {
    name: 'Borders & Inputs',
    colors: [
      { variable: '--border', label: 'Border' },
      { variable: '--input', label: 'Input' },
      { variable: '--ring', label: 'Ring' },
    ],
  },
  {
    name: 'Destructive',
    colors: [
      { variable: '--destructive', label: 'Destructive' },
      { variable: '--destructive-background', label: 'Destructive BG' },
      { variable: '--destructive-foreground', label: 'Destructive FG' },
      { variable: '--destructive-separator', label: 'Destructive Sep' },
    ],
  },
  {
    name: 'Warning',
    colors: [
      { variable: '--warning', label: 'Warning' },
      { variable: '--warning-background', label: 'Warning BG' },
      { variable: '--warning-foreground', label: 'Warning FG' },
      { variable: '--warning-separator', label: 'Warning Sep' },
    ],
  },
  {
    name: 'Success',
    colors: [
      { variable: '--success', label: 'Success' },
      { variable: '--success-background', label: 'Success BG' },
      { variable: '--success-foreground', label: 'Success FG' },
      { variable: '--success-separator', label: 'Success Sep' },
    ],
  },
  {
    name: 'Info',
    colors: [
      { variable: '--info-background', label: 'Info BG' },
      { variable: '--info-foreground', label: 'Info FG' },
      { variable: '--info-separator', label: 'Info Sep' },
    ],
  },
  {
    name: 'Chart',
    colors: [
      { variable: '--chart-1', label: 'Chart 1' },
      { variable: '--chart-2', label: 'Chart 2' },
      { variable: '--chart-3', label: 'Chart 3' },
      { variable: '--chart-4', label: 'Chart 4' },
      { variable: '--chart-5', label: 'Chart 5' },
    ],
  },
  {
    name: 'Sidebar',
    colors: [
      { variable: '--sidebar-background', label: 'Sidebar BG' },
      { variable: '--sidebar-foreground', label: 'Sidebar FG' },
      { variable: '--sidebar-primary', label: 'Sidebar Primary' },
      { variable: '--sidebar-primary-foreground', label: 'Sidebar Primary FG' },
      { variable: '--sidebar-accent', label: 'Sidebar Accent' },
      { variable: '--sidebar-accent-foreground', label: 'Sidebar Accent FG' },
      { variable: '--sidebar-border', label: 'Sidebar Border' },
      { variable: '--sidebar-ring', label: 'Sidebar Ring' },
    ],
  },
]

function ColorPalette() {
  return (
    <div className="flex flex-col gap-6">
      {colorGroups.map((group, index) => (
        <div key={group.name} className="flex flex-col gap-6">
          {index > 0 && <Separator />}
          <h3 className="text-sm font-semibold">{group.name}</h3>
          <div className="flex flex-wrap gap-4">
            {group.colors.map((color) => (
              <ColorSwatch key={color.variable} {...color} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

const meta: Meta = {
  title: 'Foundation/Colors',
}

export default meta
type Story = StoryObj

export const Palette: Story = {
  render: () => <ColorPalette />,
}
