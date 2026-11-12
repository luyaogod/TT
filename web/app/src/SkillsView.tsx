// 设置页「技能」的两个子标签：技能库（登记） / 分发（装到五个平台）。
import { useState } from 'react'
import { SkillsLibrary } from './SkillsLibrary'
import { SkillsDeploy } from './SkillsDeploy'
import { cn } from '../../shared/utils'

type Pane = 'lib' | 'deploy'

export function SkillsView() {
  const [pane, setPane] = useState<Pane>('lib')
  return (
    <div className="space-y-4">
      <div className="inline-flex items-center gap-0.5 rounded-ui border border-border p-0.5">
        {([['lib', '技能库'], ['deploy', '分发']] as const).map(([id, label]) => (
          <button key={id} type="button" onClick={() => setPane(id)}
            className={cn(
              'rounded-[4px] px-3 py-1 text-[12px] transition-colors',
              pane === id ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:text-foreground',
            )}>
            {label}
          </button>
        ))}
      </div>
      {pane === 'lib' ? <SkillsLibrary /> : <SkillsDeploy />}
    </div>
  )
}
