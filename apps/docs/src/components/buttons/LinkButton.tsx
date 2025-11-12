import { clsx as cn } from 'clsx'
import type { ComponentProps } from 'react'

import styles from './Button.module.scss'

interface Props {
  variant?: 'default' | 'ghost'
}
export function LinkButton({
  children,
  className,
  variant = 'default',
  ...props
}: Props & ComponentProps<'a'>) {
  return (
    <a
      className={cn(
        styles.button,
        {
          [styles.ghost]: variant === 'ghost',
          [styles.default]: variant === 'default',
        },
        className
      )}
      {...props}
    >
      {children}
    </a>
  )
}
