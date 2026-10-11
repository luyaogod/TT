param([int]$TargetPid, [int]$MaxNodes = 60)
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$root = [System.Windows.Automation.AutomationElement]::RootElement
$cond = New-Object System.Windows.Automation.PropertyCondition([System.Windows.Automation.AutomationElement]::ProcessIdProperty, $TargetPid)
$wins = $root.FindAll([System.Windows.Automation.TreeScope]::Children, $cond)
$out = New-Object System.Collections.Generic.List[string]
foreach ($w in $wins) {
  $walker = [System.Windows.Automation.TreeWalker]::ControlViewWalker
  $queue = New-Object System.Collections.Queue
  $queue.Enqueue(@($w, 0))
  $n = 0
  while ($queue.Count -gt 0 -and $n -lt $MaxNodes) {
    $item = $queue.Dequeue()
    $el = $item[0]; $d = $item[1]
    $r = $el.Current.BoundingRectangle
    $line = ("{0}{1} | name='{2}' | autoId='{3}' | rect={4},{5} {6}x{7}" -f ("  " * $d), $el.Current.ControlType.ProgrammaticName, $el.Current.Name, $el.Current.AutomationId, [int]$r.X, [int]$r.Y, [int]$r.Width, [int]$r.Height)
    $out.Add($line)
    $n++
    if ($d -ge 6) { continue }
    $child = $walker.GetFirstChild($el)
    while ($child -ne $null) {
      $queue.Enqueue(@($child, $d + 1))
      $child = $walker.GetNextSibling($child)
    }
  }
}
[IO.File]::WriteAllLines("$PWD\uia-tree.txt", $out, (New-Object Text.UTF8Encoding($false)))
Write-Output ("nodes=" + $out.Count)
